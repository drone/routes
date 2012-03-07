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
        fmt.Fprintf(w, "your are %s %s", lastName)
        fmt.Fprintf(w, "your are %s %s", firstName)
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

### Http HEAD
If you define a GET method for a route the HEAD method is also enabled. The response body will be stripped, and the content-length will be set to zero.

### Http OPTIONS
OPTIONS requests are automatically handled. Assume we define the following routes:

    mux.Put("/home", handler)
    mux.Post("/home", handler)

When the following HTTP request is submitted:

    OPTIONS /home HTTP/1.1

It will return the following response:

    HTTP/1.1 200 OK
    Public: PUT, POST
    Content-Length: 0

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

# LICENSE
Copyright (c) 2012 Brad Rydzewski

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
