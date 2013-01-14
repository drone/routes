package router

import (
	"bufio"
	"net"
	"net/http"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	"github.com/drone/routes/exp/context"
)

const (
	DELETE  = "DELETE"
	GET     = "GET"
	HEAD    = "HEAD"
	OPTIONS = "OPTIONS"
	PATCH   = "PATCH"
	POST    = "POST"
	PUT     = "PUT"
)

type route struct {
	method  string
	regex   *regexp.Regexp
	params  map[int]string
	handler http.HandlerFunc
}

type Router struct {
	sync.RWMutex
	routes  []*route
	filters []http.HandlerFunc
	params  map[string]interface{}
}

func New() *Router {
	r := Router{}
	r.params = make(map[string]interface{})
	return &r
}

// Get adds a new Route for GET requests.
func (r *Router) Get(pattern string, handler http.HandlerFunc) {
	r.AddRoute(GET, pattern, handler)
}

// Put adds a new Route for PUT requests.
func (r *Router) Put(pattern string, handler http.HandlerFunc) {
	r.AddRoute(PUT, pattern, handler)
}

// Del adds a new Route for DELETE requests.
func (r *Router) Del(pattern string, handler http.HandlerFunc) {
	r.AddRoute(DELETE, pattern, handler)
}

// Patch adds a new Route for PATCH requests.
func (r *Router) Patch(pattern string, handler http.HandlerFunc) {
	r.AddRoute(PATCH, pattern, handler)
}

// Post adds a new Route for POST requests.
func (r *Router) Post(pattern string, handler http.HandlerFunc) {
	r.AddRoute(POST, pattern, handler)
}

// Adds a new Route for Static http requests. Serves
// static files from the specified directory
func (r *Router) Static(pattern string, dir string) {
	//append a regex to the param to match everything
	// that comes after the prefix
	pattern = pattern + "(.+)"
	r.Get(pattern, func(w http.ResponseWriter, req *http.Request) {
		path := filepath.Clean(req.URL.Path)
		path = filepath.Join(dir, path)
		http.ServeFile(w, req, path)
	})
}

// Adds a new Route to the Handler
func (r *Router) AddRoute(method string, pattern string, handler http.HandlerFunc) {
	r.Lock()
	defer r.Unlock()

	//split the url into sections
	parts := strings.Split(pattern, "/")

	//find params that start with ":"
	//replace with regular expressions
	j := 0
	params := make(map[int]string)
	for i, part := range parts {
		if strings.HasPrefix(part, ":") {
			expr := "([^/]+)"
			//a user may choose to override the defult expression
			// similar to expressjs: ‘/user/:id([0-9]+)’ 
			if index := strings.Index(part, "("); index != -1 {
				expr = part[index:]
				part = part[:index]
			}
			params[j] = part[1:]
			parts[i] = expr
			j++
		}
	}

	//recreate the url pattern, with parameters replaced
	//by regular expressions. then compile the regex
	pattern = strings.Join(parts, "/")
	regex := regexp.MustCompile(pattern)

	route := &route{
		method  : method,
		regex   : regex,
		handler : handler,
		params  : params,
	}

	//append to the list of Routes
	r.routes = append(r.routes, route)
}

// Filter adds the middleware filter.
func (r *Router) Filter(filter http.HandlerFunc) {
	r.Lock()
	r.filters = append(r.filters, filter)
	r.Unlock()
}

// FilterParam adds the middleware filter iff the URL parameter exists.
func (r *Router) FilterParam(param string, filter http.HandlerFunc) {
	r.Filter(func(w http.ResponseWriter, req *http.Request) {
		c := context.Get(req)
		if len(c.Params.Get(param)) > 0 { filter(w, req) }
	})
}

// FilterPath adds the middleware filter iff the path matches the request.
func (r *Router) FilterPath(path string, filter http.HandlerFunc) {
	pattern := path
	pattern = strings.Replace(pattern, "*", "(.+)", -1)
	pattern = strings.Replace(pattern, "**", "([^/]+)", -1)
	regex := regexp.MustCompile(pattern)
	r.Filter(func(w http.ResponseWriter, req *http.Request) {
		if regex.MatchString(req.URL.Path) { filter(w, req) }
	})
}

// Required by http.Handler interface. This method is invoked by the
// http server and will handle all page routing
func (r *Router) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	r.RLock()
	defer r.RUnlock()

	//wrap the response writer in our custom interface
	w := &responseWriter{writer: rw, Router: r}

	//find a matching Route
	for _, route := range r.routes {

		//if the methods don't match, skip this handler
		//i.e if request.Method is 'PUT' Route.Method must be 'PUT'
		if req.Method != route.method {
			continue
		}

		//check if Route pattern matches url
		if !route.regex.MatchString(req.URL.Path) {
			continue
		}

		//get submatches (params)
		matches := route.regex.FindStringSubmatch(req.URL.Path)

		//double check that the Route matches the URL pattern.
		if len(matches[0]) != len(req.URL.Path) {
			continue
		}

		//create the http.Requests context
		c := context.Get(req)

		//add url parameters to the context
		for i, match := range matches[1:] {
			c.Params.Set(route.params[i], match)
		}

		//execute middleware filters
		for _, filter := range r.filters {
			filter(w, req)
			if w.started { return }
		}

		//invoke the request handler
		route.handler(w, req)
		return
	}

	//if no matches to url, throw a not found exception
	if w.started == false {
		http.NotFound(w, req)
	}
}

// responseWriter is a wrapper for the http.ResponseWriter to track if
// response was written to, and to store a reference to the router.
type responseWriter struct {
	Router  *Router
	writer  http.ResponseWriter
	started bool
	status  int
}

// Header returns the header map that will be sent by WriteHeader.
func (w *responseWriter) Header() http.Header {
	return w.writer.Header()
}

// Write writes the data to the connection as part of an HTTP reply,
// and sets `started` to true
func (w *responseWriter) Write(p []byte) (int, error) {
	w.started = true
	return w.writer.Write(p)
}

// WriteHeader sends an HTTP response header with status code,
// and sets `started` to true
func (w *responseWriter) WriteHeader(code int) {
	w.status = code
	w.started = true
	w.writer.WriteHeader(code)
}

// The Hijacker interface is implemented by ResponseWriters that allow an
// HTTP handler to take over the connection.
func (w *responseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return w.writer.(http.Hijacker).Hijack()
}
