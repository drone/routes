package routes

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"io/ioutil"
	"net/http"
	"strconv"
)

// Helper Functions to Read from the http.Request Body -------------------------

// ReadJson parses the JSON-encoded data in the http.Request object and
// stores the result in the value pointed to by v.
func ReadJson(r *http.Request, v interface{}) error {
	body, err := ioutil.ReadAll(r.Body)
	r.Body.Close()
	if err != nil {
		return err
	}
	return json.Unmarshal(body, v)
}

// ReadXml parses the XML-encoded data in the http.Request object and
// stores the result in the value pointed to by v.
func ReadXml(r *http.Request, v interface{}) error {
	body, err := ioutil.ReadAll(r.Body)
	r.Body.Close()
	if err != nil {
		return err
	}
	return xml.Unmarshal(body, v)
}

// Helper Functions to Write to the http.ReponseWriter -------------------------

// ServeJson writes the JSON representation of resource v to the
// http.ResponseWriter.
func ServeJson(w http.ResponseWriter, v interface{}) {
	content, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Length", strconv.Itoa(len(content)))
	w.Header().Set("Content-Type", "application/json")
	w.Write(content)
}

// ServeXml writes the XML representation of resource v to the
// http.ResponseWriter.
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

// ServeTemplate applies the named template to the specified data map and
// writes the output to the http.ResponseWriter.
func ServeTemplate(w http.ResponseWriter, name string, data map[string]interface{}) {
	// cast the writer to the resposneWriter, get the router
	r := w.(*responseWriter).Router

	r.RLock()
	defer r.RUnlock()

	if data == nil {
		data = map[string]interface{}{}
	}

	// append global params to the template
	for k, v := range r.params {
		data[k] = v
	}

	var buf bytes.Buffer
	if err := r.views.ExecuteTemplate(&buf, name, data); err != nil {
		panic(err)
		return
	}

	// set the content length, type, etc
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write(buf.Bytes())
}

// Error will terminate the http Request with the specified error code.
func Error(w http.ResponseWriter, code int) {
	http.Error(w, http.StatusText(code), code)
}
