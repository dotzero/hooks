package models

import (
	"bytes"
	"context"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRequestHugeBody(t *testing.T) {
	body := strings.Repeat(`{}`, 2048*10)
	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, "/", bytes.NewBufferString(body))
	assert.NoError(t, err)

	model := NewRequest(req)
	assert.Len(t, model.Body, maxBodySize)
}

func TestRequestContentTypeBoundary(t *testing.T) {
	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, "/", bytes.NewBufferString(""))
	assert.NoError(t, err)

	req.Header.Add("Content-Type", `multipart/form-data;boundary="boundary"`)

	model := NewRequest(req)
	assert.Equal(t, "multipart/form-data", model.ContentType)
}

func TestRequestStripHeaders(t *testing.T) {
	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, "/", bytes.NewBufferString(""))
	assert.NoError(t, err)

	req.Header.Add("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9")
	req.Header.Add("Accept-Encoding", "gzip, deflate")
	req.Header.Add("Accept-Language", "en-GB,en-US;q=0.9,en;q=0.8,ru;q=0.7")
	req.Header.Add("Cache-Control", "max-age=0")
	req.Header.Add("Host", "example.org")
	req.Header.Add("Upgrade-Insecure-Requests", "1")
	req.Header.Add("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/100.0.4896.127 Safari/537.36")
	req.Header.Add("X-Real-Ip", "0.0.0.0") // ignore
	req.Header.Add("X-Amzn-Trace-Id", "Root=1-6268e794-11cd66b95ec7e46b039ab76e")
	req.Header.Add("X-Forwarded-For", "0.0.0.0") // ignore

	model := NewRequest(req)
	assert.Len(t, model.Headers, 8)
}

func TestRequestPost(t *testing.T) {
	form := url.Values{}
	form.Set("param", "value")
	form.Add("params[]", "value1")
	form.Add("params[]", "value2")

	body := bytes.NewBufferString(form.Encode())

	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, "/", body)
	assert.NoError(t, err)

	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	model := NewRequest(req)
	assert.Len(t, model.FormData, 2)
}

func TestRequestGet(t *testing.T) {
	form := url.Values{}
	form.Set("param", "value")
	form.Add("params[]", "value1")
	form.Add("params[]", "value2")

	url := "/?" + form.Encode()

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, url, bytes.NewBufferString(""))
	assert.NoError(t, err)

	model := NewRequest(req)
	assert.Len(t, model.QueryData, 2)
}
