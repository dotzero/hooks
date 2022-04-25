package handlers

import (
	"errors"
	"net/http"
	"net/url"
	"strconv"

	"github.com/go-chi/render"
)

var errNotFound = errors.New("Hook is not found")

func renderError(w http.ResponseWriter, r *http.Request, err error) {
	switch err {
	case errNotFound:
		render.Status(r, http.StatusNotFound)
	default:
		render.Status(r, http.StatusInternalServerError)
	}

	render.JSON(w, r, map[string]string{"error": err.Error()})
}

func parseBool(form url.Values, key string) bool {
	v := form.Get(key)
	if len(v) == 0 {
		return false
	}

	result, err := strconv.ParseBool(v)
	if err != nil {
		panic(err)
	}

	return result
}
