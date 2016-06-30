package routes

import (
	"compress/flate"
	"compress/gzip"
	"encoding/json"
	"encoding/xml"
	"io/ioutil"
	"net/http"
	"net/url"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
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
)

type AcceptEncoding int

const (
	PassThrough AcceptEncoding = iota
	Gzip
)

//commonly used mime-types
const (
	applicationJson = "application/json"
	applicationXml  = "application/xml"
	textXml         = "text/xml"
)

type route struct {
	method  string
	regex   *regexp.Regexp
	params  map[int]string
	handler http.HandlerFunc
}

type RouteMux struct {
	routes  []*route
	filters []http.HandlerFunc
}

func New() *RouteMux {
	return &RouteMux{}
}

// Get adds a new Route for GET requests.
func (m *RouteMux) Get(pattern string, handler http.HandlerFunc) {
	m.AddRoute(GET, pattern, handler)
}

// Put adds a new Route for PUT requests.
func (m *RouteMux) Put(pattern string, handler http.HandlerFunc) {
	m.AddRoute(PUT, pattern, handler)
}

// Del adds a new Route for DELETE requests.
func (m *RouteMux) Del(pattern string, handler http.HandlerFunc) {
	m.AddRoute(DELETE, pattern, handler)
}

// Delete is same as Del but created for compatibility
func (m *RouteMux) Delete(pattern string, handler http.HandlerFunc) {
	m.AddRoute(DELETE, pattern, handler)
}

// Patch adds a new Route for PATCH requests.
func (m *RouteMux) Patch(pattern string, handler http.HandlerFunc) {
	m.AddRoute(PATCH, pattern, handler)
}

// Post adds a new Route for POST requests.
func (m *RouteMux) Post(pattern string, handler http.HandlerFunc) {
	m.AddRoute(POST, pattern, handler)
}

// Adds a new Route for Static http requests. Serves
// static files from the specified directory
func (m *RouteMux) Static(pattern string, dir string) {
	//append a regex to the param to match everything
	// that comes after the prefix
	pattern = pattern + "(.+)"
	m.AddRoute(GET, pattern, func(w http.ResponseWriter, r *http.Request) {
		path := filepath.Clean(r.URL.Path)
		path = filepath.Join(dir, path)
		http.ServeFile(w, r, path)
	})
}

// Adds a new Route to the Handler
func (m *RouteMux) AddRoute(method string, pattern string, handler http.HandlerFunc) {

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
	}

	//now create the Route
	route := &route{}
	route.method = method
	route.regex = regex
	route.handler = handler
	route.params = params

	//and finally append to the list of Routes
	m.routes = append(m.routes, route)
}

// Filter adds the middleware filter.
func (m *RouteMux) Filter(filter http.HandlerFunc) {
	m.filters = append(m.filters, filter)
}

// FilterParam adds the middleware filter iff the REST URL parameter exists.
func (m *RouteMux) FilterParam(param string, filter http.HandlerFunc) {
	if !strings.HasPrefix(param, ":") {
		param = ":" + param
	}

	m.Filter(func(w http.ResponseWriter, r *http.Request) {
		p := r.URL.Query().Get(param)
		if len(p) > 0 {
			filter(w, r)
		}
	})
}

// Required by http.Handler interface. This method is invoked by the
// http server and will handle all page routing
func (m *RouteMux) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	requestPath := r.URL.Path
	started := false
	// Set the trailer headers which will follow status code
	w.Header().Set("Trailer", "Content-Encoding, Content-Length, Content-Type")
	//find a matching Route
	for _, route := range m.routes {

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

		if len(route.params) > 0 {
			//add url parameters to the query param map
			values := r.URL.Query()
			for i, match := range matches[1:] {
				values.Add(route.params[i], match)
			}

			//reassemble query params and add to RawQuery
			r.URL.RawQuery = url.Values(values).Encode() + "&" + r.URL.RawQuery
			//r.URL.RawQuery = url.Values(values).Encode()
		}

		//execute middleware filters
		for _, filter := range m.filters {
			filter(w, r)
			if started {
				return
			}
		}

		//Invoke the request handler
		route.handler(w, r)
		started = true
		break
	}

	//if no matches to url, throw a not found exception
	if started == false {
		http.NotFound(w, r)
	}
}

// -----------------------------------------------------------------------------
// Below are helper functions to replace boilerplate
// code that serializes resources and writes to the
// http response.

// ServeJson replies to the request with a JSON
// representation of resource v.
func ServeJson(w http.ResponseWriter, v interface{}) {
	w.Header().Set("Content-Type", applicationJson)
	err := json.NewEncoder(w).Encode(v)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	return
}

//ServeJSONEncode is eager JSON writer
// with gzip encoding where possible
func ServeJSONEncode(w http.ResponseWriter, r *http.Request, v interface{}) {
	w.Header().Set("Content-Type", applicationJson)
	var pipe AcceptEncoding
	// Accept-Encoding has gzip
	if strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
		pipe = Gzip
	} else {
		pipe = PassThrough
	}

	switch pipe {
	case PassThrough:
		err := json.NewEncoder(w).Encode(v)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	case Gzip:
		w.Header().Set("Content-Encoding", "gzip")
		gz, err := gzip.NewWriterLevel(w, flate.DefaultCompression)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer gz.Close()
		json.NewEncoder(gz).Encode(v)
	}
	return
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
	w.Header().Set("Content-Length", strconv.Itoa(len(content)))
	w.Header().Set("Content-Type", "text/xml; charset=utf-8")
	w.Write(content)
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
