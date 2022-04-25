package models

import (
	"io/ioutil"
	"mime"
	"net/http"
	"time"
)

// Request is a hook request model
type Request struct {
	Name          string
	RemoteAddr    string
	Method        string
	Path          string
	Query         string
	Body          string
	ContentType   string
	ContentLength int64
	Headers       map[string]string
	FormData      map[string]string
	QueryData     map[string]string
	Created       time.Time
}

// NewRequest returns a new request model
func NewRequest(r *http.Request) *Request {
	body, _ := ioutil.ReadAll(r.Body)
	defer func() { _ = r.Body.Close() }()

	contentType := r.Header.Get("Content-Type")

	d, _, err := mime.ParseMediaType(contentType)
	if err == nil && d != "" {
		contentType = d
	}

	req := &Request{
		Name:          tinyID(),
		RemoteAddr:    r.RemoteAddr,
		Method:        r.Method,
		Path:          r.URL.Path,
		Query:         r.URL.RawQuery,
		Body:          string(body),
		ContentType:   contentType,
		ContentLength: r.ContentLength,
		Headers:       make(map[string]string, len(r.Header)),
		FormData:      make(map[string]string, len(r.PostForm)),
		QueryData:     make(map[string]string, len(r.URL.Query())),
		Created:       time.Now(),
	}

	for name, value := range r.Header {
		req.Headers[name] = value[0]
	}

	for name, value := range r.PostForm {
		req.FormData[name] = value[0]
	}

	for name, value := range r.URL.Query() {
		req.QueryData[name] = value[0]
	}

	return req
}
