package handlers

import (
	"errors"
	"io"
	"net/http"
	"testing"

	"github.com/go-chi/chi"
	"github.com/stretchr/testify/assert"

	"github.com/dotzero/hooks/app/models"
)

func TestWebHome(t *testing.T) {
	s := &storeMock{
		RecentHooksFunc: func(max int) ([]*models.Hook, error) {
			return nil, nil
		},
	}
	tmpl := &tplMock{
		ExecuteFunc: func(wr io.Writer, data interface{}) error {
			return nil
		},
	}

	handler := WebHome(s, tmpl, "", 48)

	router := chi.NewRouter()
	router.Get("/", handler)

	w, err := testRequest(router, http.MethodGet, "/", "")

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestWebHomeError(t *testing.T) {
	s := &storeMock{
		RecentHooksFunc: func(max int) ([]*models.Hook, error) {
			return nil, nil
		},
	}
	tmpl := &tplMock{
		ExecuteFunc: func(wr io.Writer, data interface{}) error {
			return errors.New("storage error")
		},
	}

	handler := WebHome(s, tmpl, "", 48)

	router := chi.NewRouter()
	router.Get("/", handler)

	w, err := testRequest(router, http.MethodGet, "/", "")

	assert.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.JSONEq(t, `{"error":"storage error"}`, w.Body.String())
}

func TestWebInspect(t *testing.T) {
	s := &storeMock{
		HookFunc: func(name string) (*models.Hook, error) {
			assert.Equal(t, "foo", name)

			return &models.Hook{
				Name: name,
			}, nil
		},
		RecentHooksFunc: func(max int) ([]*models.Hook, error) {
			return nil, nil
		},
		RequestsFunc: func(hook string) ([]*models.Request, error) {
			assert.Equal(t, "foo", hook)

			return nil, nil
		},
	}
	tmpl := &tplMock{
		ExecuteFunc: func(wr io.Writer, data interface{}) error {
			return nil
		},
	}

	handler := WebInspect(s, tmpl, "", 48)

	router := chi.NewRouter()
	router.Get("/{hook}", handler)

	w, err := testRequest(router, http.MethodGet, "/foo", "")

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestWebInspectPrivate(t *testing.T) {
	s := &storeMock{
		HookFunc: func(name string) (*models.Hook, error) {
			assert.Equal(t, "private", name)

			return &models.Hook{
				Name:    name,
				Private: true,
				Secret:  "private", // stored in testRequest()
			}, nil
		},
		RecentHooksFunc: func(max int) ([]*models.Hook, error) {
			return nil, nil
		},
		RequestsFunc: func(hook string) ([]*models.Request, error) {
			assert.Equal(t, "private", hook)

			return nil, nil
		},
	}
	tmpl := &tplMock{
		ExecuteFunc: func(wr io.Writer, data interface{}) error {
			return nil
		},
	}

	handler := WebInspect(s, tmpl, "", 48)

	router := chi.NewRouter()
	router.Get("/{hook}", handler)

	w, err := testRequest(router, http.MethodGet, "/private", "")

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestWebInspectPrivateError(t *testing.T) {
	s := &storeMock{
		HookFunc: func(name string) (*models.Hook, error) {
			assert.Equal(t, "foo", name)

			return &models.Hook{
				Name:    name,
				Private: true,
				Secret:  "private", // stored in testRequest()
			}, nil
		},
	}
	tmpl := &tplMock{
		ExecuteFunc: func(wr io.Writer, data interface{}) error {
			return nil
		},
	}

	handler := WebInspect(s, tmpl, "", 48)

	router := chi.NewRouter()
	router.Get("/{hook}", handler)

	w, err := testRequest(router, http.MethodGet, "/foo", "")

	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, w.Code) // name - secret missmatch
}
