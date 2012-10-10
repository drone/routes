# routes.go
a simple http routing API for the Go programming language

    go get github.com/drone/routes

for more information see:
http://gopkgdoc.appspot.com/pkg/github.com/bradrydzewski/routes

[![](https://drone.io/drone/routes/status.png)](https://drone.io/drone/routes/latest)

## Getting Started

    package main

    import (
        "fmt"
        "github.com/drone/routes"
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

    mux.Get("/:param", handler)
    mux.Put("/:param", handler)
    mux.Post("/:param", handler)
    mux.Patch("/:param", handler)
    mux.Del("/:param", handler)

You can specify custom regular expressions for routes:

    mux.Get("/files/:param(.+)", handler)

You can also create routes for static files:

    pwd, _ := os.Getwd()
    mux.Static("/static", pwd)

this will serve any files in `/static`, including files in subdirectories. For example `/static/logo.gif` or `/static/style/main.css`.

## Filters / Middleware
You can apply filters to routes, which is useful for enforcing security,
redirects, etc.

You can, for example, filter all request to enforce some type of security:

    var FilterUser = func(w http.ResponseWriter, r *http.Request) {
    	if r.URL.User == nil || r.URL.User.Username() != "admin" {
    		http.Error(w, "", http.StatusUnauthorized)
    	}
    }

    r.Filter(FilterUser)

You can also apply filters only when certain REST URL Parameters exist:

    r.Get("/:id", handler)
    r.Filter("id", func(rw http.ResponseWriter, r *http.Request) {
		...
	})

## Helper Functions
You can use helper functions for serializing to Json and Xml. I found myself constantly writing code to serialize, set content type, content length, etc. Feel free to use these functions to eliminate redundant code in your app.

Helper function for serving Json, sets content type to `application/json`:

    func handler(w http.ResponseWriter, r *http.Request) {
		mystruct := { ... }
        routes.ServeJson(w, &mystruct)
    }

Helper function for serving Xml, sets content type to `application/xml`:

    func handler(w http.ResponseWriter, r *http.Request) {
		mystruct := { ... }
        routes.ServeXml(w, &mystruct)
    }

Helper function to serve Xml OR Json, depending on the value of the `Accept` header:

    func handler(w http.ResponseWriter, r *http.Request) {
		mystruct := { ... }
        routes.ServeFormatted(w, r, &mystruct)
    }

