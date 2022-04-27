package app

import (
	"fmt"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/cors"
	"github.com/go-chi/render"

	"github.com/dotzero/hooks/app/handlers"
)

func (a *App) makeHTTPServer(address string, port int, router http.Handler) *http.Server {
	return &http.Server{
		Addr:              fmt.Sprintf("%s:%d", address, port),
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
		IdleTimeout:       30 * time.Second,
	}
}

func (a *App) routes() *chi.Mux {
	router := chi.NewRouter()

	corsMiddleware := cors.New(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"HEAD", "GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		MaxAge:         300,
	})

	router.Use(middleware.Logger)
	router.Use(middleware.NoCache)
	router.Use(middleware.RealIP)
	router.Use(middleware.Recoverer)
	router.Use(middleware.RedirectSlashes)
	router.Use(corsMiddleware.Handler)

	router.Get("/", handlers.WebHome(a.Storage, a.Templates.Lookup("home.html"), a.AppURL, a.BoltTTL))
	router.Get("/i/{hook}", handlers.WebInspect(a.Storage, a.Templates.Lookup("hook.html"), a.AppURL, a.BoltTTL))
	router.Post("/api/create", handlers.APICreate(a.Storage))
	router.Get("/api/stats", handlers.APIStats(a.Storage, a.BoltTTL))
	router.Handle("/{hook}", handlers.APIHook(a.Storage))

	router.Get("/robots.txt", func(w http.ResponseWriter, r *http.Request) {
		render.PlainText(w, r, "User-agent: *\n")
	})

	router.Get("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, filepath.Join("static", "/favicon.ico"))
	})

	// file server for static content from /static
	fileServer(router, "/static", http.Dir("static"))

	return router
}

func fileServer(r chi.Router, path string, root http.FileSystem) {
	origPath := path
	fs := http.StripPrefix(path, http.FileServer(root))

	if path != "/" && path[len(path)-1] != '/' {
		r.Get(path, http.RedirectHandler(path+"/", http.StatusMovedPermanently).ServeHTTP)
		path += "/"
	}

	path += "*"

	r.Get(path, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// don't show dirs, just serve files
		if strings.HasSuffix(r.URL.Path, "/") && len(r.URL.Path) > 1 && r.URL.Path != (origPath+"/") {
			http.NotFound(w, r)
			return
		}
		fs.ServeHTTP(w, r)
	}))
}
