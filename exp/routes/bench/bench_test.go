package bench

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/drone/routes/exp/routes"
	gorilla "code.google.com/p/gorilla/mux"
	"github.com/bmizerany/pat"
)

func HandlerOk(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "hello world")
	w.WriteHeader(http.StatusOK)
}

// Benchmark_Routes runs a benchmark against our custom Mux using the
// default settings.
func Benchmark_Routes(b *testing.B) {

	handler := routes.NewRouter()
	handler.Get("/person/:last/:first", HandlerOk)

	for i := 0; i < b.N; i++ {
		r, _ := http.NewRequest("GET", "/person/anderson/thomas?learn=kungfu", nil)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, r)
	}
}

// Benchmark_Web runs a benchmark against the pat.go Mux using the
// default settings.
func Benchmark_Pat(b *testing.B) {

	m := pat.New()
	m.Get("/person/:last/:first", http.HandlerFunc(HandlerOk))

	for i := 0; i < b.N; i++ {
		r, _ := http.NewRequest("GET", "/person/anderson/thomas?learn=kungfu", nil)
		w := httptest.NewRecorder()
		m.ServeHTTP(w, r)
	}
}

// Benchmark_Gorilla runs a benchmark against the Gorilla Mux using
// the default settings.
func Benchmark_GorillaHandler(b *testing.B) {

	handler := gorilla.NewRouter()
	handler.HandleFunc("/person/{last}/{first}", HandlerOk)

	for i := 0; i < b.N; i++ {
		r, _ := http.NewRequest("GET", "/person/anderson/thomas?learn=kungfu", nil)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, r)
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
		mux.ServeHTTP(w, r)
	}
}
