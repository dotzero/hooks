package handlers

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/go-chi/chi"
	"github.com/stretchr/testify/assert"

	"github.com/dotzero/hooks/app/models"
)

func TestAPICreate(t *testing.T) {
	s := &storeMock{
		PutHookFunc: func(hook *models.Hook) error {
			assert.True(t, hook.Private)

			return nil
		},
	}

	handler := APICreate(s)

	router := chi.NewRouter()
	router.Post("/api", handler)

	form := url.Values{}
	form.Set("private", "true")

	w, err := testRequest(router, http.MethodPost, "/api", form.Encode())

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"private":true`)
}

func TestAPIHook(t *testing.T) {
	s := &storeMock{
		HookFunc: func(name string) (*models.Hook, error) {
			assert.Equal(t, "foo", name)

			return &models.Hook{
				Name: name,
			}, nil
		},
		PutRequestFunc: func(hook string, req *models.Request) error {
			assert.Equal(t, "foo", hook)
			assert.Equal(t, http.MethodPost, req.Method)
			assert.Equal(t, "foo=bar&foobar", req.Query)

			return nil
		},
	}

	handler := APIHook(s)

	router := chi.NewRouter()
	router.Handle("/{hook}", handler)

	w, err := testRequest(router, http.MethodPost, "/foo?foo=bar&foobar", "")

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, w.Code)
}

func testRequest(h http.Handler, method string, address string, body string) (*httptest.ResponseRecorder, error) {
	req, err := http.NewRequestWithContext(
		context.Background(),
		method,
		address,
		bytes.NewBuffer([]byte(body)),
	)
	if err != nil {
		return nil, err
	}

	if method == http.MethodPost {
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	}

	resp := httptest.NewRecorder()
	h.ServeHTTP(resp, req)

	return resp, nil
}
