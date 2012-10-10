/*
Package routes a simple http routing API for the Go programming language,
compatible with the standard http.ListenAndServe function.

Create a new route multiplexer:

	r := routes.NewRouter()

Define a simple route with a given method (ie Get, Put, Post ...), path and
http.HandleFunc.

	r.Get("/foo", fooHandler)

Define a route with restful parameters in the path:

	r.Get("/:foo/:bar", func(rw http.ResponseWriter, req *http.Request) {
		c := routes.NewContext(req)
		foo := c.Params.Get("foo")
		bar := c.Params.Get("bar")
		fmt.Fprintf(rw, "%s %s", foo, bar)
	})

The parameters are parsed from the URL, and stored in the Request Context.

More control over the route's parameter matching is possible by providing
a custom regular expression:

	r.Get("/files/:file(.+)", handler)

To start the web server, use the standard http.ListenAndServe
function, and provide the route multiplexer:

    http.Handle("/", r)
    http.ListenAndServe(":8000", nil)

*/
package routes
