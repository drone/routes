# routes.go
a simple http routing API for the Go programming language

    go get github.com/drone/routes

for more information see:
http://gopkgdoc.appspot.com/pkg/github.com/drone/routes

[![](https://drone.io/drone/routes/status.png)](https://drone.io/drone/routes/latest)

## Getting Started

    package main

    import (
        "fmt"
        "github.com/drone/routes"
        "net/http"
    )

    func foobar (w http.ResponseWriter, r *http.Request) {
		c := routes.NewContext(r)
		foo := c.Params.Get(":foo")
		bar := c.Params.Get(":bar")
		fmt.Fprintf(w, "%s %s", foo, bar)
    }

    func main() {
        r := routes.NewRouter()
        r.Get("/:bar/:foo", foobar)

        http.Handle("/", r)
        http.ListenAndServe(":8088", nil)
    }

### Route Examples
You can create routes for all http methods:

    r.Get("/:param", handler)
    r.Put("/:param", handler)
    r.Post("/:param", handler)
    r.Patch("/:param", handler)
    r.Del("/:param", handler)

You can specify custom regular expressions for routes:

    r.Get("/files/:param(.+)", handler)

You can also create routes for static files:

    pwd, _ := os.Getwd()
    r.Static("/static", pwd)

this will serve any files in `/static`, including files in subdirectories. For
example `/static/logo.gif` or `/static/style/main.css`.

## Filters / Middleware
You can implement route filters to do things like enforce security, set session
variables, etc

You can, for example, filter all request to enforce some type of security:

    r.Filter(func(rw http.ResponseWriter, r *http.Request) {
    	if r.URL.User != "admin" {
    		http.Error(w, "", http.StatusForbidden)
    	}
	})

You can also apply filters only when certain REST URL Parameters exist:

    r.Get("/:id", handler)
    r.Filter("id", func(rw http.ResponseWriter, r *http.Request) {
		c := routes.NewContext(r)
		id := c.Params.Get("id")

		// verify the user has access to the specified resource id
    	user := r.URL.User.Username()
	    if HasAccess(user, id) == false {
			http.Error(w, "", http.StatusForbidden)
		}
	})

## Helper Functions
You can use helper functions for serializing to Json and Xml. I found myself
constantly writing code to serialize, set content type, content length, etc.
Feel free to use these functions to eliminate redundant code in your app.

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
