# routes.go
a simple http routing API for the Go programming language

    go get github.com/bradrydzewski/routes.go

## Getting Started

    package main

    import (
        "fmt"
        "github.com/bradrydzewski/routes.go"
        "net/http"
    )

    func Whoami(w http.ResponseWriter, r *http.Request) {
        params := r.URL.Query()
        lastName := params.Get(":last")
        firstName := params.Get(":first")
        fmt.Fprintf(w, "you are %s %s", firstName, lastName)
    }

    func main() {
        mux := routes.New()
        mux.Get("/:last/:first", Whoami)

        http.Handle("/", mux)
        http.ListenAndServe(":8088", nil)
    }

### Route Examples
You can create routes for all http methods:

    mux.Get("/:param1", handler)
    mux.Put("/:param1", handler)
    mux.Post("/:param1", handler)
    mux.Patch("/:param1", handler)
    mux.Del("/:param1", handler)

You can also create routes for static files:

    pwd, _ := os.Getwd()
    mux.Static("/static", pwd)

this will serve any files in `/static`, including files in subdirectories. For example `/static/logo.gif` or `/static/style/main.css`.

## Helper Functions
You can use helper functions for serializing to Json and Xml. I found myself constantly writing code to serialize, set content type, content length, etc. Feel free to use these functions to eliminate boilerplate code in your app.

Helper function for serving Json, sets content type to `application/json`:

    func handler(w http.ResponseWriter, r *http.Request) {
        routes.ServeJson(w, &mystruct)
    }

Helper function for serving Xml, sets content type to `application/xml`:

    func handler(w http.ResponseWriter, r *http.Request) {
        routes.ServeXml(w, &mystruct)
    }

Helper function to serve Xml OR Json, depending on the value of the `Accept` header:

    func handler(w http.ResponseWriter, r *http.Request) {
        routes.ServeXml(w, &mystruct)
    }

## Logging
Logging is enabled by default, but can be disabled:

    mux.Logging = false

You can also specify your logger:

    mux.Logger = log.New(os.Stdout, "", 0)
