package routes

import (
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"text/template"
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
	views   *template.Template
	params  map[string]interface{}
}

func NewRouter() *Router {
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
	regex, regexErr := regexp.Compile(pattern)
	if regexErr != nil {
		panic(regexErr)
	}

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
		c := NewContext(req)
		if len(c.Params.Get(param)) > 0 { filter(w, req) }
	})
}

// Set stores the specified key / value pair.
func (r *Router) Set(name string, value interface{}) {
	r.Lock()
	r.params[name] = value
	r.Unlock()
}

// SetEnv stores the specified environment variable as a key / value pair. If
// the environment variable is not set the default value will be used
func (r *Router) SetEnv(name, value string) {
	r.Lock()
	defer r.Unlock()

	env := os.Getenv(name)
	if len(env) == 0 { env = value }
	r.Set(name, env)
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
		c := NewContext(req)

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

// Template uses the provided template definitions.
func (r *Router) Template(t *template.Template) {
	r.Lock()
	defer r.Unlock()
	r.views = template.Must(t.Clone())
}

// TemplateFiles parses the template definitions from the named files.
func (r *Router) TemplateFiles(filenames ...string) {
	r.Lock()
	defer r.Unlock()
	r.views = template.Must(template.ParseFiles(filenames...))
}

// TemplateGlob parses the template definitions from the files identified
// by the pattern, which must match at least one file.
func (r *Router) TemplateGlob(pattern string) {
	r.Lock()
	defer r.Unlock()
	r.views = template.Must(template.ParseGlob(pattern))
}
