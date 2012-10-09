# routes.go
a simple http routing API for the Go programming language

    go get github.com/drone/routes

this project combines the best of `web.go` and `pat.go`. It uses `pat.go`'s named url parameters (ie `:param`) and `web.go`'s regular expression groups for url matching and parameter extraction, which provides significant performance improvements.

for more information see:
http://gopkgdoc.appspot.com/pkg/github.com/bradrydzewski/routes

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

## Security
You can restrict access to routes by assigning an `AuthHandler` to a route.

Here is an example using a custom `AuthHandler` per route. Image we are doing some type of Basic authentication:

    func authHandler(w http.ResponseWriter, r *http.Request) bool {
	    user := r.URL.User.Username()
	    password := r.URL.User.Password()
	    if user != "xxx" && password != "xxx" {
            // if we wanted, we could do an http.Redirect here
		    return false
	    }
	    return true
    }

    mux.Get("/:param", handler).SecureFunc(authHandler)

If you plan to use the same `AuthHandler` to secure all of your routes, you may want to set the `DefaultAuthHandler`:

    routes.DefaulAuthHandler = authHandler
    mux.Get("/:param", handler).Secure()
    mux.Get("/:param", handler).Secure()

### OAuth2
In the above examples, we implemented our own custom `AuthHandler`. Check out the [auth.go](https://github.com/bradrydzewski/auth.go) API which provides custom AuthHandlers for OAuth2 providers such as Google and Github.

## Logging
Logging is enabled by default, but can be disabled:

    mux.Logging = false

You can also specify your logger:

    mux.Logger = log.New(os.Stdout, "", 0)
