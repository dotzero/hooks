package handlers

import (
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/render"

	"github.com/dotzero/hooks/app/models"
)

// APICreate handle the creation of new hooks
func APICreate(s store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			renderError(w, r, err)
			return
		}

		private := parseBool(r.Form, "private")
		hook := models.NewHook(private)

		if err := s.PutHook(hook); err != nil {
			renderError(w, r, err)
			return
		}

		if private {
			cookie := &http.Cookie{
				Name:     "hook_" + hook.Name,
				Value:    hook.Secret,
				Path:     "/",
				MaxAge:   86400, // 24 hours
				HttpOnly: false,
			}
			http.SetCookie(w, cookie)
		}

		render.JSON(w, r, hook)
	}
}

// APIHook handle requests to hooks
func APIHook(s store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			renderError(w, r, err)
			return
		}

		hook, err := s.Hook(chi.URLParam(r, "hook"))
		if err != nil {
			renderError(w, r, err)
			return
		}

		if hook == nil {
			renderError(w, r, errNotFound)
			return
		}

		req := models.NewRequest(r)

		if err := s.PutRequest(hook.Name, req); err != nil {
			renderError(w, r, err)
			return
		}

		render.JSON(w, r, req)
	}
}
