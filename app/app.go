package app

import (
	"fmt"
	"html/template"
	"image/color"
	"log"
	"os"
	"path/filepath"

	"github.com/dustin/go-humanize"
	"github.com/zero-pkg/tpl"

	"github.com/dotzero/hooks/app/favicon"
	"github.com/dotzero/hooks/app/storage"
)

// App is a Hook application
type App struct {
	Storage   *storage.BoltDB
	Templates *tpl.Templates

	CommonOpts
}

const (
	boltFile = "hooks.db"
)

// New prepares application
func New(commonOpts CommonOpts) (*App, error) {
	app := &App{}
	app.SetCommon(commonOpts)

	if err := app.setupDataStore(); err != nil {
		return nil, err
	}

	if err := app.setupTpl(); err != nil {
		return nil, err
	}

	return app, nil
}

// Run listens on the TCP network address srv.Addr and then
// calls Serve to handle requests on incoming connections.
func (a *App) Run(address string, port int) error {
	addr := fmt.Sprintf("%s:%d", address, port)
	log.Printf("[INFO] http server listen at: http://" + addr)

	return a.makeHTTPServer(address, port, a.routes()).ListenAndServe()
}

// setupDataStore initializes BoltDB store
func (a *App) setupDataStore() error {
	if err := makeDirs(a.BoltPath); err != nil {
		return err
	}

	store, err := storage.New(filepath.Join(a.BoltPath, boltFile))
	if err != nil {
		return err
	}

	a.Storage = store

	return nil
}

// setupTpl initializes Templates engine
func (a *App) setupTpl() error {
	funcMap := template.FuncMap{
		"safeURL": func(u string) template.URL {
			return template.URL(u) // nolint
		},
		"rgb": func(rbga [4]uint8) string {
			return fmt.Sprintf("%d, %d, %d", rbga[0], rbga[1], rbga[2])
		},
		"favicon": func(rbga [4]uint8) string {
			return favicon.New(favicon.WithColor(color.RGBA{
				rbga[0], rbga[1], rbga[2], rbga[3],
			})).String()
		},
		"humanizeTime": humanize.Time,
		"humanizeSize": func(size int64) string {
			return humanize.Bytes(uint64(size))
		},
	}

	templ, err := tpl.New().Funcs(funcMap).ParseDir(a.TmlPath, a.TplExt)
	if err != nil {
		return err
	}

	a.Templates = templ

	return nil
}

func makeDirs(dirs ...string) error {
	// exists returns whether the given file or directory exists or not
	exists := func(path string) (bool, error) {
		_, err := os.Stat(path)
		if err == nil {
			return true, nil
		}

		if os.IsNotExist(err) {
			return false, nil
		}

		return true, err
	}

	for _, dir := range dirs {
		ex, err := exists(dir)
		if err != nil {
			return fmt.Errorf("can't check directory status for %s", dir)
		}

		if !ex {
			if e := os.MkdirAll(dir, 0700); e != nil { // nolint
				return fmt.Errorf("can't make directory %s", dir)
			}
		}
	}

	return nil
}
