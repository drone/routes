package routes

import (
	"io"
	"net/http"
)

// Context stores data for the duration of the http.Request
type Context struct {
	// named parameters that are passed in via RESTful URL Parameters
	Params Params

	// named attributes that persist for the lifetime of the request
	Values Values

	// reference to the parent http.Request
	req *http.Request
}

// Retruns the Context associated with the http.Request.
func NewContext(r *http.Request) *Context {

	// get the context bound to the http.Request
	if v, ok := r.Body.(*wrapper); ok {
		return v.context
	}

	// create a new context
	c := Context{ }
	c.Params = make(Params)
	c.Values = make(Values)
	c.req = r

	// wrap the request and bind the context
	wrapper := wrap(r)
	wrapper.context = &c
	return &c
}

// Retruns the parent http.Request to which the context is bound.
func (c *Context) Request() *http.Request {
	return c.req
}

// wrapper decorates an http.Request's Body (io.ReadCloser) so that we can
// bind a Context to the Request. This is obviously a hack that i'd rather
// avoid, however, it is for the greater good ...
//
// NOTE: If this turns out to be a really stupid approach we can use this
//       approach from the go mailing list: http://goo.gl/Vw13f which I
//       avoided because I didn't want a global lock
type wrapper struct {
	body    io.ReadCloser // the original message body
	context *Context
}

func wrap(r *http.Request) *wrapper {
	w := wrapper{ body: r.Body }
	r.Body = &w
	return &w
}

func (w *wrapper) Read(p []byte) (n int, err error) {
	return w.body.Read(p)
}

func (w *wrapper) Close() error {
	return w.body.Close()
}

// Parameter Map ---------------------------------------------------------------

// Params maps a string key to a list of values.
type Params map[string]string

// Get gets the first value associated with the given key. If there are
// no values associated with the key, Get returns the empty string.
func (p Params) Get(key string) string {
	if p == nil {
		return ""
	}
	return p[key]
}

// Set sets the key to value. It replaces any existing values.
func (p Params) Set(key, value string) {
	p[key] = value
}

// Del deletes the values associated with key.
func (p Params) Del(key string) {
	delete(p, key)
}

// Value Map -------------------------------------------------------------------

// Values maps a string key to a list of values.
type Values map[interface{}]interface{}

// Get gets the value associated with the given key. If there are
// no values associated with the key, Get returns nil.
func (v Values) Get(key interface{}) interface{} {
	if v == nil {
		return nil
	}

	return v[key]
}

// GetStr gets the value associated with the given key in string format.
// If there are no values associated with the key, Get returns an
// empty string.
func (v Values) GetStr(key interface{}) interface{} {
	if v == nil { return "" }
	
	val := v.Get(key)
	if val == nil { return "" }

	str, ok := val.(string)
	if !ok { return "" }
	return str
}

// Set sets the key to value. It replaces any existing values.
func (v Values) Set(key, value interface{}) {
	v[key] = value
}

// Del deletes the values associated with key.
func (v Values) Del(key interface{}) {
	delete(v, key)
}
