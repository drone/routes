package routes

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
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
		t.Errorf("Code set to [%s]; want [%s]", w.Code, http.StatusNotFound)
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

// TestHead tests the ability to serve a
// request for a HEAD method
func TestHead(t *testing.T) {

	r, _ := http.NewRequest("HEAD", "/routes_test.go", nil)
	w := httptest.NewRecorder()
	pwd, _ := os.Getwd()

	handler := new(RouteMux)
	handler.Static("/", pwd)
	handler.ServeHTTP(w, r)

	if w.Body.String() != "" {
		t.Errorf("a head method should have a zero-length body")
	}

	contentLength := w.Header().Get("Content-Length")
	if contentLength != "0" {
		t.Errorf("Content-Lenght set to [%s]; want [%s]", contentLength, "0")
	}
}

// TestOptions tests the ability to handle
// an HTTP OPTIONS request
func TestOptions(t *testing.T) {

	r, _ := http.NewRequest("OPTIONS", "/", nil)
	w := httptest.NewRecorder()

	handler := new(RouteMux)
	handler.Put("/", HandlerOk)
	handler.Post("/", HandlerOk)
	handler.Get("/", HandlerOk)
	handler.ServeHTTP(w, r)

	options := w.Header().Get("Public")
	optionsExpected := "GET, HEAD, OPTIONS, POST, PUT"
	if options != optionsExpected {
		t.Errorf("Options set to [%s]; want [%s]", options, optionsExpected)
	}
}

// Benchmark_RoutedHandler runs a benchmark against
// the RouteMux using the default settings.
func Benchmark_RoutedHandler(b *testing.B) {

	r, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler := new(RouteMux)
	handler.Get("/", HandlerOk)

	for i := 0; i < b.N; i++ {
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
