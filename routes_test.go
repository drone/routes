package routes

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"
)

var HandlerOk = func(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "hello world")
	w.WriteHeader(http.StatusOK)
}

var HandlerEmpty = func(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

var HandlerLargeResponseServeJSONEncode = func(w http.ResponseWriter, r *http.Request) {
	i := []int{}
	for j := 1; j <= 4000; j++ {
		i = append(i, j)
	}
	w.WriteHeader(http.StatusOK)
	ServeJSONEncode(w, r, i)
}

var HandlerDoubleWriteStatus = func(w http.ResponseWriter, r *http.Request) {
	i := []int{}
	for j := 1; j <= 4000; j++ {
		i = append(i, j)
	}
	w.WriteHeader(http.StatusCreated)
	ServeJSONEncode(w, r, i)

}

var HandlerLargeResponseServeJSON = func(w http.ResponseWriter, r *http.Request) {
	i := []int{}
	for j := 1; j <= 4000; j++ {
		i = append(i, j)
	}
	w.WriteHeader(http.StatusOK)
	ServeJson(w, i)
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

// TestDeleteMethod tests. Test compatibility
// to http spec with method names

func TestRouteDeleteMethod(t *testing.T) {
	r, _ := http.NewRequest("DELETE", "/person/anderson/thomas?learn=kungfu", nil)
	w := httptest.NewRecorder()

	handler := new(RouteMux)
	handler.Delete("/person/:last/:first", HandlerEmpty)
	handler.ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("Basic check for DELETE event failed")
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

func TestHttpStatusIssue(t *testing.T) {
	r, _ := http.NewRequest("GET", "/", nil)
	r.Header.Set("Accept-Encoding", "gzip")
	w := httptest.NewRecorder()
	handler := new(RouteMux)
	handler.Get("/", HandlerDoubleWriteStatus)
	handler.ServeHTTP(w, r)
	if w.Code != http.StatusCreated {
		t.Errorf("Code set to [%v]; want [%v]. Header error", w.Code, http.StatusCreated)
	}
	if w.Header().Get("Content-Encoding") != "gzip" {
		t.Errorf("Code trailer header Content-Type not present")
	}
}

//TestPostServeJsonEncode test gzip encoding protocol within POST calls
func TestPostServeJsonEncode(t *testing.T) {
	form := url.Values{}
	form.Add("s", "8")
	form.Add("h", "5")
	form.Add("r", "4")
	r, _ := http.NewRequest("POST", "/", strings.NewReader(form.Encode()))
	r.Header.Set("Accept-Encoding", "gzip")
	w := httptest.NewRecorder()

	handler := new(RouteMux)
	handler.Post("/", HandlerLargeResponseServeJSONEncode)
	handler.ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("Code set to [%v]; want [%v]", w.Code, http.StatusOK)
	}
	i := []int{}
	for j := 1; j <= 4000; j++ {
		i = append(i, j)
	}
	content, _ := json.Marshal(i)
	if w.Header().Get("Content-Encoding") != "gzip" {
		t.Errorf("Data is not encoded with gzip")
	}
	body, _ := ioutil.ReadAll(w.Body)
	if len(body) >= len(content) {
		t.Errorf("Issues with gzip encoding content-length: %d and json length %d", len(body), len(content))
	}
}

//TestServeJsonEncode eager gzip encoding of JSON
// whereever possible.

//TestServeJsonEncode eager gzip encoding of JSON
// whereever possible.

func TestServeJsonEncode(t *testing.T) {

	r, _ := http.NewRequest("GET", "/", nil)
	r.Header.Set("Accept-Encoding", "gzip")
	w := httptest.NewRecorder()

	handler := new(RouteMux)
	handler.Get("/", HandlerLargeResponseServeJSONEncode)
	handler.ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("Code set to [%v]; want [%v]", w.Code, http.StatusOK)
	}
	i := []int{}
	for j := 1; j <= 4000; j++ {
		i = append(i, j)
	}
	content, _ := json.Marshal(i)
	if w.Header().Get("Content-Encoding") != "gzip" {
		t.Errorf("Data is not encoded with gzip")
	}
	body, _ := ioutil.ReadAll(w.Body)
	if len(body) >= len(content) {
		t.Errorf("Issues with gzip encoding content-length: %d and json length %d", len(body), len(content))
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

// Benchmark ServeJson for performance
func Benchmark_ServeJson(b *testing.B) {
	r, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	mux := http.NewServeMux()
	mux.HandleFunc("/", HandlerLargeResponseServeJSON)

	for i := 0; i < b.N; i++ {
		mux.ServeHTTP(w, r)

	}
}

// Benchmark ServeJsonEncode for performance
func Benchmark_ServeJsonEncode(b *testing.B) {
	r, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	mux := http.NewServeMux()
	mux.HandleFunc("/", HandlerLargeResponseServeJSONEncode)

	for i := 0; i < b.N; i++ {
		mux.ServeHTTP(w, r)

	}
}

// Benchmark ServeJsonEncode for performance
func Benchmark_ServeJsonEncodeGzip(b *testing.B) {
	r, _ := http.NewRequest("GET", "/", nil)
	r.Header.Set("Accept-Encoding", "gzip")
	w := httptest.NewRecorder()

	mux := http.NewServeMux()
	mux.HandleFunc("/", HandlerLargeResponseServeJSONEncode)

	for i := 0; i < b.N; i++ {
		mux.ServeHTTP(w, r)
		w.Flush()

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
