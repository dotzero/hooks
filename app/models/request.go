package models

import (
	"io"
	"io/ioutil"
	"mime"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	maxBodySize = 1024 * 10
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
	_ = r.ParseForm()

	return &Request{
		Name:          tinyID(),
		RemoteAddr:    r.RemoteAddr,
		Method:        r.Method,
		Path:          r.URL.Path,
		Query:         r.URL.RawQuery,
		Body:          parseBody(r.Body),
		ContentType:   parseContentType(r.Header),
		ContentLength: r.ContentLength,
		Headers:       parseHeaders(r.Header),
		FormData:      parseFormData(r.PostForm),
		QueryData:     parseQueryData(r.URL.Query()),
		Created:       time.Now(),
	}
}

func parseBody(reader io.ReadCloser) string {
	body, _ := ioutil.ReadAll(reader)
	defer func() { _ = reader.Close() }()

	if len(body) > maxBodySize {
		return string(body[0:maxBodySize])
	}

	return string(body)
}

func parseContentType(headers http.Header) string {
	contentType := headers.Get("Content-Type")

	d, _, err := mime.ParseMediaType(contentType)
	if err == nil && d != "" {
		return d
	}

	return contentType
}

func parseHeaders(headers http.Header) map[string]string {
	ignore := map[string]struct{}{
		"x-varnish":                {},
		"x-forwarded-for":          {},
		"x-heroku-dynos-in-use":    {},
		"x-request-start":          {},
		"x-heroku-queue-wait-time": {},
		"x-heroku-queue-depth":     {},
		"x-real-ip":                {},
		"x-forwarded-proto":        {},
		"x-via":                    {},
		"x-forwarded-port":         {},
	}

	parsed := make(map[string]string, len(headers))

	for name, value := range headers {
		if _, ok := ignore[strings.ToLower(name)]; !ok {
			parsed[name] = value[0]
		}
	}

	return parsed
}

func parseFormData(form url.Values) map[string]string {
	parsed := make(map[string]string, len(form))

	for name, value := range form {
		parsed[name] = value[0]
	}

	return parsed
}

func parseQueryData(query url.Values) map[string]string {
	parsed := make(map[string]string, len(query))

	for name, value := range query {
		parsed[name] = value[0]
	}

	return parsed
}
