package routes

import (
	"bufio"
	"net"
	"net/http"
)

// ResponseWriter is a wrapper for the http.ResponseWriter to track if
// response was written to.
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
