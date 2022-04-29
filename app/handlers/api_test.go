package handlers

import (
	"errors"
	"net/http"
	"net/url"
	"strings"
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

func TestAPICreateError(t *testing.T) {
	s := &storeMock{
		PutHookFunc: func(hook *models.Hook) error {
			return errors.New("storage error")
		},
	}

	handler := APICreate(s)

	router := chi.NewRouter()
	router.Post("/api", handler)

	form := url.Values{}
	form.Set("private", "true")

	w, err := testRequest(router, http.MethodPost, "/api", form.Encode())

	assert.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.JSONEq(t, `{"error":"storage error"}`, w.Body.String())
}

func TestAPIHookCases(t *testing.T) {
	cases := []struct {
		name      string // test case name
		method    string
		query     string
		body      string
		formData  map[string]string
		queryData map[string]string
	}{
		{
			name:   "get",
			method: http.MethodGet,
			query:  "?param=value&single",
			queryData: map[string]string{
				"param":  "value",
				"single": "",
			},
		},
		{
			name:   "post",
			method: http.MethodPost,
			body:   "private=true",
			formData: map[string]string{
				"private": "true",
			},
		},
		{
			name:   "json",
			method: http.MethodPost,
			body:   `{"private": true}`,
		},
		{
			name:   "put",
			method: http.MethodPut,
			body:   `{"private": true}`,
		},
	}

	for _, c := range cases {
		c := c // pin
		t.Run(c.name, func(t *testing.T) {
			s := &storeMock{
				HookFunc: func(name string) (*models.Hook, error) {
					assert.Equal(t, "foo", name)

					return &models.Hook{
						Name: name,
					}, nil
				},
				PutRequestFunc: func(hook string, req *models.Request) error {
					assert.Equal(t, "foo", hook)
					assert.Equal(t, c.method, req.Method)
					assert.Equal(t, strings.TrimPrefix(c.query, "?"), req.Query)

					switch req.Headers["Content-Type"] {
					case "application/json":
						assert.Equal(t, c.body, req.Body)
					case "application/x-www-form-urlencoded":
						assert.Equal(t, c.formData, req.FormData)
					default:
						assert.Equal(t, c.queryData, req.QueryData)
					}

					return nil
				},
			}

			handler := APIHook(s)

			router := chi.NewRouter()
			router.Handle("/{hook}", handler)

			w, err := testRequest(router, c.method, "/foo"+c.query, c.body)

			assert.NoError(t, err)
			assert.Equal(t, http.StatusOK, w.Code)
		})
	}
}

func TestAPIStats(t *testing.T) {
	s := &storeMock{
		CountFunc: func(name []byte) (int, error) {
			return 10, nil
		},
	}

	handler := APIStats(s, 24)

	router := chi.NewRouter()
	router.Post("/api", handler)

	w, err := testRequest(router, http.MethodPost, "/api", "")

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.JSONEq(t, `{"hooks":10,"ttl_hours":24}`, w.Body.String())
}

func TestAPIStatsError(t *testing.T) {
	s := &storeMock{
		CountFunc: func(name []byte) (int, error) {
			return 0, errors.New("storage error")
		},
	}

	handler := APIStats(s, 24)

	router := chi.NewRouter()
	router.Post("/api", handler)

	w, err := testRequest(router, http.MethodPost, "/api", "")

	assert.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.JSONEq(t, `{"error":"storage error"}`, w.Body.String())
}
