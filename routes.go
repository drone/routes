package routes

import (
	"encoding/json"
	"encoding/xml"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const (
	CONNECT = "CONNECT"
	DELETE  = "DELETE"
	GET     = "GET"
	HEAD    = "HEAD"
	OPTIONS = "OPTIONS"
	PATCH   = "PATCH"
	POST    = "POST"
	PUT     = "PUT"
	TRACE   = "TRACE"

	// log format, modeled after http://wiki.nginx.org/HttpLogModule
	LOG = `%s - - [%s] "%s %s %s" %d %d "%s" "%s"`

	// commonly used mime types
	applicationJson = "application/json"
	applicationXml  = "applicatoin/xml"
	textXml         = "text/xml"
)

type Route struct {
	method  string
	regex   *regexp.Regexp
	params  map[int]string
	handler http.HandlerFunc
	auth    AuthHandler
}

type RouteMux struct {
	routes  []*Route
	Logging bool
	Logger  *log.Logger
}

func New() *RouteMux {
	routeMux := RouteMux{}
	routeMux.Logging = true
	routeMux.Logger = log.New(os.Stdout, "", 0)
	return &routeMux
}

// Adds a new Route to the Handler
func (this *RouteMux) AddRoute(method string, pattern string, handler http.HandlerFunc) *Route {

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
			params[j] = part
			parts[i] = expr
			j++
		}
	}

	//recreate the url pattern, with parameters replaced
	//by regular expressions. then compile the regex
	pattern = strings.Join(parts, "/")
	regex, regexErr := regexp.Compile(pattern)
	if regexErr != nil {
		//TODO add error handling here to avoid panic
		panic(regexErr)
		return nil
	}

	//now create the Route
	route := &Route{}
	route.method = method
	route.regex = regex
	route.handler = handler
	route.params = params

	//and finally append to the list of Routes
	this.routes = append(this.routes, route)

	return route
}

// Adds a new Route for GET requests
func (this *RouteMux) Get(pattern string, handler http.HandlerFunc) *Route {
	return this.AddRoute(GET, pattern, handler)
}

// Adds a new Route for PUT requests
func (this *RouteMux) Put(pattern string, handler http.HandlerFunc) *Route {
	return this.AddRoute(PUT, pattern, handler)
}

// Adds a new Route for DELETE requests
func (this *RouteMux) Del(pattern string, handler http.HandlerFunc) *Route {
	return this.AddRoute(DELETE, pattern, handler)
}

// Adds a new Route for PATCH requests
// See http://www.ietf.org/rfc/rfc5789.txt
func (this *RouteMux) Patch(pattern string, handler http.HandlerFunc) *Route {
	return this.AddRoute(PATCH, pattern, handler)
}

// Adds a new Route for POST requests
func (this *RouteMux) Post(pattern string, handler http.HandlerFunc) *Route {
	return this.AddRoute(POST, pattern, handler)
}

// Secures a route using the default AuthHandler
func (this *Route) Secure() *Route {
	this.auth = DefaultAuthHandler
	return this
}

// SecureFunc a route using a custom AuthHandler function
func (this *Route) SecureFunc(handler AuthHandler) *Route {
	this.auth = handler
	return this
}

// Adds a new Route for Static http requests. Serves
// static files from the specified directory
func (this *RouteMux) Static(pattern string, dir string) *Route {
	//append a regex to the param to match everything
	// that comes after the prefix
	pattern = pattern + "(.+)"
	return this.AddRoute(GET, pattern, func(w http.ResponseWriter, r *http.Request) {
		path := filepath.Clean(r.URL.Path)
		path = filepath.Join(dir, path)
		http.ServeFile(w, r, path)
	})
}

// Required by http.Handler interface. This method is invoked by the
// http server and will handle all page routing
func (this *RouteMux) ServeHTTP(rw http.ResponseWriter, r *http.Request) {

	requestPath := r.URL.Path

	//wrap the response writer, in our custom interface
	w := &responseWriter{writer: rw}

	//find a matching Route
	for _, route := range this.routes {

		//if the methods don't match, skip this handler
		//i.e if request.Method is 'PUT' Route.Method must be 'PUT'
		if r.Method != route.method {
			continue
		}

		//check if Route pattern matches url
		if !route.regex.MatchString(requestPath) {
			continue
		}

		//get submatches (params)
		matches := route.regex.FindStringSubmatch(requestPath)

		//double check that the Route matches the URL pattern.
		if len(matches[0]) != len(requestPath) {
			continue
		}

		//add url parameters to the query param map
		values := r.URL.Query()
		for i, match := range matches[1:] {
			values.Add(route.params[i], match)
		}

		//reassemble query params and add to RawQuery
		r.URL.RawQuery = url.Values(values).Encode()

		//enfore security, if necessary
		if route.auth != nil {
			//autenticate the user
			ok := route.auth(w, r)
			//if the auth handler redirected the user
			//or already wrote a response, we can just exit
			if w.started {
				return
			} else if ok == false {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
			}
		}

		//Invoke the request handler
		route.handler(w, r)

		break
	}

	//if no matches to url, throw a not found exception
	if w.started == false {
		http.NotFound(w, r)
	}

	//if logging is turned on
	if this.Logging {
		this.Logger.Printf(LOG, r.RemoteAddr, time.Now().String(), r.Method,
			r.URL.Path, r.Proto, w.status, w.size,
			r.Referer(), r.UserAgent())
	}
}

// ---------------------------------------------------------------------------------
// Simple wrapper around a ResponseWriter

//responseWriter is a wrapper for the http.ResponseWriter
// to track if response was written to. It also allows us
// to automatically set certain headers, such as Content-Type,
// Access-Control-Allow-Origin, etc.
type responseWriter struct {
	writer  http.ResponseWriter // Writer
	started bool
	size    int
	status  int
}

// Header returns the header map that will be sent by WriteHeader.
func (this *responseWriter) Header() http.Header {
	return this.writer.Header()
}

// Write writes the data to the connection as part of an HTTP reply,
// and sets `started` to true
func (this *responseWriter) Write(p []byte) (int, error) {
	this.size += len(p)
	this.started = true
	return this.writer.Write(p)
}

// WriteHeader sends an HTTP response header with status code,
// and sets `started` to true
func (this *responseWriter) WriteHeader(code int) {
	this.status = code
	this.started = true
	this.writer.WriteHeader(code)
}

// ---------------------------------------------------------------------------------
// Authentication helper functions to enable user authentication

type AuthHandler func(http.ResponseWriter, *http.Request) bool

// DefaultAuthHandler will be applied to any route when the Secure() function
// is invoked, as opposed to SecureFunc(), which accepts a custom AuthHandler.
//
// By default, the DefaultAuthHandler will deny all requests. This value
// should be replaced with a custom AuthHandler implementation, as this
// is just a dummy function.
var DefaultAuthHandler = func(w http.ResponseWriter, r *http.Request) bool {
	return false
}

// ---------------------------------------------------------------------------------
// Below are helper functions to replace boilerplate
// code that serializes resources and writes to the
// http response.

// ServeJson replies to the request with a JSON
// representation of resource v.
func ServeJson(w http.ResponseWriter, v interface{}) {
	content, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write(content)
	w.Header().Set("Content-Length", strconv.Itoa(len(content)))
	w.Header().Set("Content-Type", applicationJson)
}

// ReadJson will parses the JSON-encoded data in the http
// Request object and stores the result in the value
// pointed to by v.
func ReadJson(r *http.Request, v interface{}) error {
	body, err := ioutil.ReadAll(r.Body)
	r.Body.Close()
	if err != nil {
		return err
	}
	return json.Unmarshal(body, v)
}

// ServeXml replies to the request with an XML
// representation of resource v.
func ServeXml(w http.ResponseWriter, v interface{}) {
	content, err := xml.Marshal(v)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write(content)
	w.Header().Set("Content-Length", strconv.Itoa(len(content)))
	w.Header().Set("Content-Type", "text/xml; charset=utf-8")
}

// ReadXml will parses the XML-encoded data in the http
// Request object and stores the result in the value
// pointed to by v.
func ReadXml(r *http.Request, v interface{}) error {
	body, err := ioutil.ReadAll(r.Body)
	r.Body.Close()
	if err != nil {
		return err
	}
	return xml.Unmarshal(body, v)
}

// ServeFormatted replies to the request with
// a formatted representation of resource v, in the
// format requested by the client specified in the
// Accept header.
func ServeFormatted(w http.ResponseWriter, r *http.Request, v interface{}) {
	accept := r.Header.Get("Accept")
	switch accept {
	case applicationJson:
		ServeJson(w, v)
	case applicationXml, textXml:
		ServeXml(w, v)
	default:
		ServeJson(w, v)
	}

	return
}
