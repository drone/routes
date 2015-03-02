package routes

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"
)

var HandlerOk = func(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "hello world")
	w.WriteHeader(http.StatusOK)
}

var HandlerErr = func(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "", http.StatusBadRequest)
}

var FilterUser = func(w http.ResponseWriter, r *http.Request) {
	if r.URL.User == nil || r.URL.User.Username() != "admin" {
		http.Error(w, "", http.StatusUnauthorized)
	}
}

var FilterId = func(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get(":id")
	if id == "admin" {
		http.Error(w, "", http.StatusUnauthorized)
	}
}

// Test data used for ServeFormatted
type testData struct {
	SomeField string
	Number    int
	Test      float64
}

// contentTest describes a test for ServerFormatted
type contentTest struct {
	Header   string
	Expected string
}

// The correct formatted data for use by TestContentFormatted
var dataJSON = "{\n  \"SomeField\": \"Hi\",\n  \"Number\": 20,\n  \"Test\": 2.5\n}"
var dataXML = "<testData><SomeField>Hi</SomeField><Number>20</Number><Test>2.5</Test></testData>"

// Various test cases for TestContentFormatted
var contentTester = []contentTest{
	contentTest{
		"",
		dataJSON,
	},
	contentTest{
		"application/json",
		dataJSON,
	},
	contentTest{
		"application/xml",
		dataXML,
	},
	contentTest{
		"text/xml",
		dataXML,
	},
	contentTest{
		"text/plain; q=0.5, text/html,\ntext/xml; q=0.8, text/x-c",
		dataXML,
	},
	contentTest{
		"text/html;\ntext/json;",
		dataJSON,
	},
}

// TestAuthOk tests that an Auth handler will append the
// username and password to to the request URL, and will
// continue processing the request by invoking the handler.
func TestRouteOk(t *testing.T) {

	r, _ := http.NewRequest("GET", "/person/anderson/thomas?learn=kungfu", nil)
	w := httptest.NewRecorder()

	handler := new(RouteMux)
	handler.Get("/person/:last/:first", HandlerOk)
	handler.ServeHTTP(w, r)

	lastNameParam := r.URL.Query().Get(":last")
	firstNameParam := r.URL.Query().Get(":first")
	learnParam := r.URL.Query().Get("learn")

	if lastNameParam != "anderson" {
		t.Errorf("url param set to [%s]; want [%s]", lastNameParam, "anderson")
	}
	if firstNameParam != "thomas" {
		t.Errorf("url param set to [%s]; want [%s]", firstNameParam, "thomas")
	}
	if learnParam != "kungfu" {
		t.Errorf("url param set to [%s]; want [%s]", learnParam, "kungfu")
	}
}

// TestNotFound tests that a 404 code is returned in the
// response if no route matches the request url.
func TestNotFound(t *testing.T) {

	r, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	handler := new(RouteMux)
	handler.ServeHTTP(w, r)

	if w.Code != http.StatusNotFound {
		t.Errorf("Code set to [%v]; want [%v]", w.Code, http.StatusNotFound)
	}
}

// TestStatic tests the ability to serve static
// content from the filesystem
func TestStatic(t *testing.T) {

	r, _ := http.NewRequest("GET", "/routes_test.go", nil)
	w := httptest.NewRecorder()
	pwd, _ := os.Getwd()

	handler := new(RouteMux)
	handler.Static("/", pwd)
	handler.ServeHTTP(w, r)

	testFile, _ := ioutil.ReadFile(pwd + "/routes_test.go")
	if w.Body.String() != string(testFile) {
		t.Errorf("handler.Static failed to serve file")
	}
}

// TestFilter tests the ability to apply middleware function
// to filter all routes
func TestFilter(t *testing.T) {

	r, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	handler := new(RouteMux)
	handler.Get("/", HandlerOk)
	handler.Filter(FilterUser)
	handler.ServeHTTP(w, r)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Did not apply Filter. Code set to [%v]; want [%v]", w.Code, http.StatusUnauthorized)
	}

	r, _ = http.NewRequest("GET", "/", nil)
	r.URL.User = url.User("admin")
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("Code set to [%v]; want [%v]", w.Code, http.StatusOK)
	}
}

// TestFilterParam tests the ability to apply middleware
// function to filter all routes with specified parameter
// in the REST url
func TestFilterParam(t *testing.T) {

	r, _ := http.NewRequest("GET", "/:id", nil)
	w := httptest.NewRecorder()

	// first test that the param filter does not trigger
	handler := new(RouteMux)
	handler.Get("/", HandlerOk)
	handler.Get("/:id", HandlerOk)
	handler.FilterParam("id", FilterId)
	handler.ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("Code set to [%v]; want [%v]", w.Code, http.StatusOK)
	}

	// now test the param filter does trigger
	r, _ = http.NewRequest("GET", "/admin", nil)
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, r)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Did not apply Param Filter. Code set to [%v]; want [%v]", w.Code, http.StatusUnauthorized)
	}

}

// TestContentFormatted tests the basic Accept-Header content matching.
func TestContentFormatted(t *testing.T) {
	recorder := httptest.NewRecorder()
	response, _ := http.NewRequest("GET", "localhost/does/not/matter", nil)

	data := testData{
		SomeField: "Hi",
		Number:    20,
		Test:      2.5,
	}

	for _, test := range contentTester {
		response.Header.Set("Accept", test.Header)
		ServeFormatted(recorder, response, data)
		if recorder.Body.String() != test.Expected {
			t.Fatalf("Sent %q header, expected %q. Got %q", test.Header, test.Expected, recorder.Body.String())
		}
		recorder.Body.Reset()
	}

}

// Benchmark_RoutedHandler runs a benchmark against
// the RouteMux using the default settings.
func Benchmark_RoutedHandler(b *testing.B) {
	handler := new(RouteMux)
	handler.Get("/", HandlerOk)

	for i := 0; i < b.N; i++ {
		r, _ := http.NewRequest("GET", "/", nil)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, r)
	}
}

// Benchmark_RoutedHandler runs a benchmark against
// the RouteMux using the default settings with REST
// URL params.
func Benchmark_RoutedHandlerParams(b *testing.B) {

	handler := new(RouteMux)
	handler.Get("/:user", HandlerOk)

	for i := 0; i < b.N; i++ {
		r, _ := http.NewRequest("GET", "/admin", nil)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, r)
	}
}

// Benchmark_ServeMux runs a benchmark against
// the ServeMux Go function. We use this to determine
// performance impact of our library, when compared
// to the out-of-the-box Mux provided by Go.
func Benchmark_ServeMux(b *testing.B) {

	r, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	mux := http.NewServeMux()
	mux.HandleFunc("/", HandlerOk)

	for i := 0; i < b.N; i++ {
		mux.ServeHTTP(w, r)
	}
}
