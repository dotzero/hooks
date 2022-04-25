package handlers

import (
	"io"
	"net/http"
	"testing"

	"github.com/go-chi/chi"
	"github.com/stretchr/testify/assert"

	"github.com/dotzero/hooks/app/models"
)

func TestWebHome(t *testing.T) {
	s := &storeMock{}
	tmpl := &tplMock{
		ExecuteFunc: func(wr io.Writer, data interface{}) error {
			return nil
		},
	}

	handler := WebHome(s, tmpl, "")

	router := chi.NewRouter()
	router.Get("/", handler)

	w, err := testRequest(router, http.MethodGet, "/", "")

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestWebInspect(t *testing.T) {
	s := &storeMock{
		HookFunc: func(name string) (*models.Hook, error) {
			assert.Equal(t, "foo", name)

			return &models.Hook{
				Name: name,
			}, nil
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

	handler := WebInspect(s, tmpl, "")

	router := chi.NewRouter()
	router.Get("/{hook}", handler)

	w, err := testRequest(router, http.MethodGet, "/foo", "")

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, w.Code)
}
