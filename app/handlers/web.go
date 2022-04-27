package handlers

import (
	"net/http"

	"github.com/go-chi/chi"

	"github.com/dotzero/hooks/app/models"
	"github.com/dotzero/hooks/app/views"
)

const (
	maxRecent    = 10
	urlParam     = "hook"
	cookiePrefix = "hook_"
)

// WebHome handle home page
func WebHome(s store, t tpl, baseURL string, ttl int) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		recent, err := s.RecentHooks(maxRecent)
		if err != nil {
			renderError(w, r, err)
			return
		}

		err = t.Execute(w, &views.Home{
			Common: views.Common{
				BaseURL: baseURL,
				TTL:     ttl,
				Recent:  recent,
			},
		})
		if err != nil {
			renderError(w, r, err)
		}
	}
}

// WebInspect handle hook page
func WebInspect(s store, t tpl, baseURL string, ttl int) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		hook, err := s.Hook(chi.URLParam(r, urlParam))
		if err != nil {
			renderError(w, r, err)
			return
		}

		if hook == nil {
			renderError(w, r, errNotFound)
			return
		}

		if !checkAccess(r, hook) {
			renderError(w, r, errNotFound)
			return
		}

		requests, err := s.Requests(hook.Name)
		if err != nil {
			renderError(w, r, err)
			return
		}

		hook.Requests = requests

		recent, err := s.RecentHooks(maxRecent)
		if err != nil {
			renderError(w, r, err)
			return
		}

		err = t.Execute(w, &views.Hook{
			Common: views.Common{
				BaseURL: baseURL,
				TTL:     ttl,
				Recent:  recent,
			},
			Hook: hook,
		})
		if err != nil {
			renderError(w, r, err)
		}
	}
}

func checkAccess(r *http.Request, hook *models.Hook) bool {
	if !hook.Private {
		return true
	}

	c, err := r.Cookie(cookiePrefix + hook.Name)
	if c == nil || err != nil {
		return false
	}

	return c.Value == hook.Secret
}
