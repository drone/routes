/*
Package routes a simple http routing API for the Go programming language,
compatible with the standard http.ListenAndServe function.

Create a new route multiplexer:

	mux := routes.New()

Define a simple route with a given method (ie Get, Put, Post ...), path and
http.HandleFunc.

	mux.Get("/foo", fooHandler)

Define a route with restful parameters in the path:

	mux.Get("/:foo/:bar", func(w http.ResponseWriter, r *http.Request) {
		params := r.URL.Query()
		foo := params.Get(":foo")
		bar := params.Get(":bar")
		fmt.Fprintf(w, "%s %s", foo, bar)
	})

The parameters are parsed from the URL, and appended to the Request URL's
query parameters.

More control over the route's parameter matching is possible by providing
a custom regular expression:

	mux.Get("/files/:file(.+)", handler)

To start the web server, use the standard http.ListenAndServe
function, and provide the route multiplexer:

    http.Handle("/", mux)
    http.ListenAndServe(":8000", nil)

*/
package routes
