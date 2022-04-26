package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/hashicorp/logutils"
	flags "github.com/jessevdk/go-flags"

	"github.com/dotzero/hooks/app"
)

var (
	// Version is the version number or commit hash
	// These variables should be set by the linker when compiling
	Version = "0.0.0-unknown"
	// CommitHash is the git hash of last commit
	CommitHash = "Unknown"
	// CompileDate is the date of build
	CompileDate = "Unknown"
)

// Opts with command line flags and env
type Opts struct {
	Host       string `long:"host" env:"HOOKS_HOST" default:"0.0.0.0" description:"listening address"`
	Port       int    `long:"port" env:"HOOKS_PORT" default:"8080" description:"listening port"`
	AppURL     string `long:"url" env:"HOOKS_URL" default:"http://0.0.0.0:8080" description:"url to app"`
	BoltPath   string `long:"bolt-path" env:"BOLT_PATH" default:"./var" description:"parent directory for the bolt files"`
	BoltTTL    int    `long:"bolt-ttl" env:"BOLT_TTL_HOURS" default:"48" description:"TTL in hours to keep data"`
	StaticPath string `long:"static-path" env:"STATIC_PATH" default:"./static" description:"path to website assets"`
	TmlPath    string `long:"tpl-path" env:"TPL_PATH" default:"./templates" description:"path to templates files"`
	TplExt     string `long:"tpl-ext" env:"TPL_EXT" default:".html" description:"templates files extensions"`
	Verbose    bool   `long:"verbose" description:"verbose logging"`
	Version    bool   `short:"v" long:"version" description:"show the version number"`
}

func main() {
	var opts Opts

	p := flags.NewParser(&opts, flags.Default)
	if _, err := p.ParseArgs(os.Args[1:]); err != nil {
		os.Exit(1)
	}

	if opts.Version {
		fmt.Printf("Version: %s\nCommit hash: %s\nCompile date: %s\n", Version, CommitHash, CompileDate)
		os.Exit(0)
	}

	setupLog(opts.Verbose)
	log.Printf("[DEBUG] opts: %+v", opts)

	app, err := app.New(app.CommonOpts{
		AppURL:     opts.AppURL,
		BoltPath:   opts.BoltPath,
		BoltTTL:    opts.BoltTTL,
		StaticPath: opts.StaticPath,
		TmlPath:    opts.TmlPath,
		TplExt:     opts.TplExt,
	})
	if err != nil {
		log.Fatalf("[ERROR] failed to setup application, %+v", err)
	}

	if err := app.Run(context.Background(), opts.Host, opts.Port); err != nil {
		log.Fatalf("[WARN] http server terminated, %s", err)
	}
}

func setupLog(verbose bool) {
	filter := &logutils.LevelFilter{
		Levels:   []logutils.LogLevel{"DEBUG", "INFO", "WARN", "ERROR"},
		MinLevel: logutils.LogLevel("INFO"),
		Writer:   os.Stdout,
	}

	if verbose {
		log.SetFlags(log.LstdFlags | log.Lshortfile)

		filter.MinLevel = logutils.LogLevel("DEBUG")
	}

	log.SetOutput(filter)
}
