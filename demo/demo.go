package main

import (
	"fmt"
	"github.com/bradrydzewski/routes"
	"net/http"
)

func Whoami(w http.ResponseWriter, r *http.Request) {
	params := r.URL.Query()
	lastName := params.Get(":last")
	firstName := params.Get(":first")
	fmt.Fprintf(w, "your are %s %s", firstName, lastName)
}

func main() {
	mux := routes.New()
	mux.Get("/:last/:first", Whoami)

	http.Handle("/", mux)
	http.ListenAndServe(":8080", nil)
}
