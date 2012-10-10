package routes

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func HandlerOk(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "hello world")
	w.WriteHeader(http.StatusOK)
}

func HandlerSetVar(w http.ResponseWriter, r *http.Request) {
	c := NewContext(r)
	c.Values.Set("password", "z1on")
}

func HandlerErr(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "", http.StatusBadRequest)
}

// TestRouteOk tests that the route is correctly handled, and the URL parameters
// are added to the Context.
func TestRouteOk(t *testing.T) {

	r, _ := http.NewRequest("GET", "/person/anderson/thomas?learn=kungfu", nil)
	w := httptest.NewRecorder()

	mux := NewRouter()
	mux.Get("/person/:last/:first", HandlerOk)
	mux.ServeHTTP(w, r)

	c := NewContext(r)
	lastNameParam := c.Params.Get("last")
	firstNameParam := c.Params.Get("first")

	if lastNameParam != "anderson" {
		t.Errorf("url param set to [%s]; want [%s]", lastNameParam, "anderson")
	}
	if firstNameParam != "thomas" {
		t.Errorf("url param set to [%s]; want [%s]", firstNameParam, "thomas")
	}
	if w.Body.String() != "hello world" {
		t.Errorf("Body set to [%s]; want [%s]", w.Body.String(), "hello world")
	}
}

// TestFilter tests that a route is filtered prior to handling
func TestRouteFilter(t *testing.T) {

	r, _ := http.NewRequest("GET", "/person/anderson/thomas?learn=kungfu", nil)
	w := httptest.NewRecorder()

	mux := NewRouter()
	mux.Filter(HandlerSetVar)
	mux.Get("/person/:last/:first", HandlerOk)
	mux.ServeHTTP(w, r)

	c := NewContext(r)
	password := c.Values.Get("password")

	if password != "z1on" {
		t.Errorf("session variable set to [%s]; want [%s]", password, "z1on")
	}
	if w.Body.String() != "hello world" {
		t.Errorf("Body set to [%s]; want [%s]", w.Body.String(), "hello world")
	}
}

// TestFilterHalt tests that a route is filtered prior to handling, and then
// halts execution (by writing to the response).
func TestRouteFilterHalt(t *testing.T) {
	r, _ := http.NewRequest("GET", "/person/anderson/thomas?learn=kungfu", nil)
	w := httptest.NewRecorder()

	mux := NewRouter()
	mux.Filter(HandlerErr)
	mux.Get("/person/:last/:first", HandlerOk)
	mux.ServeHTTP(w, r)

	if w.Code != 400 {
		t.Errorf("Code set to [%s]; want [%s]", w.Code, http.StatusBadRequest)
	}
	if w.Body.String() == "hello world" {
		t.Errorf("Body set to [%s]; want empty", w.Body.String())
	}
}

// TestParam tests the Parameter filter, and ensures the filter is only
// executed when the specified Parameter exists.
func TestParam(t *testing.T) {
	// in the first test scenario, the Parameter filter should not
	// be triggered because the "codename" variab does not exist
	r, _ := http.NewRequest("GET", "/neo", nil)
	w := httptest.NewRecorder()

	mux := NewRouter()
	mux.Filter(HandlerSetVar)
	mux.FilterParam("codename", HandlerErr)
	mux.Get("/:nickname", HandlerOk)
	mux.ServeHTTP(w, r)

	if w.Body.String() != "hello world" {
		t.Errorf("Body set to [%s]; want [%s]", w.Body.String(), "hello world")
	}

	// in this second scenario, the Parameter filter SHOULD fire, and should
	// halt the request
	w = httptest.NewRecorder()

	mux = NewRouter()
	mux.Filter(HandlerSetVar)
	mux.FilterParam("codename", HandlerErr)
	mux.Get("/:codename", HandlerOk)
	mux.ServeHTTP(w, r)

	if w.Body.String() == "hello world" {
		t.Errorf("Body set to [%s]; want empty", w.Body.String())
	}
	if w.Code != 400 {
		t.Errorf("Code set to [%s]; want [%s]", w.Code, http.StatusBadRequest)
	}
}

/*
// TestTemplate tests template rendering
func TestTemplate(t *testing.T) {

	w := httptest.NewRecorder()

	tmpl, _ := template.New("template.html").Parse("<html><head><title>{{ .Title }}</title><body>{{ .Name }}</body></html>")

	mux := NewRouter()
	mux.Template(tmpl)
	mux.Set("Title", "Matrix")
	mux.ExecuteTemplate(w, "template.html", map[string]interface{}{ "Name" : "Morpheus" })

	if w.Body.String() != "<html><head><title>Matrix</title><body>Morpheus</body></html>" {
		t.Errorf("template not rendered correctly [%s]", w.Body.String())
	}
}
*/

// TestNotFound tests that a 404 code is returned in the
// response if no route matches the request url.
func TestNotFound(t *testing.T) {

	r, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	mux := NewRouter()
	mux.ServeHTTP(w, r)

	if w.Code != http.StatusNotFound {
		t.Errorf("Code set to [%s]; want [%s]", w.Code, http.StatusNotFound)
	}
}

// Benchmark_Routes runs a benchmark against our custom Mux using the
// default settings.
func Benchmark_Routes(b *testing.B) {

	r, _ := http.NewRequest("GET", "/person/anderson/thomas?learn=kungfu", nil)
	w := httptest.NewRecorder()
	mux := NewRouter()
	mux.Get("/person/:last/:first", HandlerOk)

	for i := 0; i < b.N; i++ {
		mux.ServeHTTP(w, r)
	}
}

// Benchmark_ServeMux runs a benchmark against the ServeMux Go function.
// We use this to determine performance impact of our library, when compared
// to the out-of-the-box Mux provided by Go.
func Benchmark_ServeMux(b *testing.B) {

	r, _ := http.NewRequest("GET", "/person/anderson/thomas?learn=kungfu", nil)
	w := httptest.NewRecorder()
	mux := http.NewServeMux()
	mux.HandleFunc("/", HandlerOk)

	for i := 0; i < b.N; i++ {
		r.URL.Query().Get("learn")
		mux.ServeHTTP(w, r)
	}
}
