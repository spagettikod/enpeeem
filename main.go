package main

import (
	"enpeeem/config"
	"enpeeem/handle"
	"enpeeem/storage"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
)

var (
	addr         string
	cfg          config.Config
	fetchAll     bool
	indexAll     bool
	indexPkg     string
	metadir      string
	pkgthreads   int
	printVersion bool
	progress     bool
	proxystash   bool
	registry     string
	urltemplate  string
	storageDir   string
	verbose      bool
	version      = "SET VERSION IN MAKEFILE"
)

func init() {
	flag.StringVar(&addr, "addr", ":8080", "network address of local registry")
	flag.StringVar(&registry, "registry", "https://registry.npmjs.org", "remote npm registry to use when the flag proxystash is set")
	flag.BoolVar(&indexAll, "index-all", false, "index all packages")
	flag.BoolVar(&progress, "progress", false, "show progress where applicable")
	flag.BoolVar(&printVersion, "version", false, "print version")
	flag.BoolVar(&verbose, "verbose", false, "print debug information")
	flag.BoolVar(&fetchAll, "fetch-all", false, "download all tarbal versions at once if a tarball is not found locally")
	flag.StringVar(&indexPkg, "index", "", "index with given package URI, example registry.npmjs.org/@types/react")
	flag.BoolVar(&proxystash, "proxystash", false, "run in proxy mode to proxy and download tarballs if not available locally")
	flag.StringVar(&metadir, "metadir", "", "metadata file directory, by default files are stored together with the tarballs")
	flag.StringVar(&urltemplate, "urltemplate", "", "Go template to rewrite tarball URL's in package metadata requests")
	flag.IntVar(&pkgthreads, "pkgthreads", 5, "number of packages to process at the same time when indexing all packages")
	flag.Usage = printUsage
}

func printUsage() {
	fmt.Printf(`Local npm registry and proxy.
	
Packages are served from the given path. Run in proxy mode to download from
remote registry and save tarballs if they are not found locally at path.

Usage:
  enpeeem [flags] <path>	

Flags:
`)
	flag.PrintDefaults()
	os.Exit(1)
}

func parseArgs() {
	flag.Parse()
	args := flag.Args()
	if printVersion {
		return
	}
	if len(args) != 1 {
		fmt.Println("error: too few arguments")
		printUsage()
		os.Exit(1)
	}
	if fetchAll && !proxystash {
		fmt.Println("info: flag fetch-all is useless without the proxystash flag")
	}
	storageDir = args[0]
}

func middleware(handler func(w http.ResponseWriter, r *http.Request) (int, error)) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		status, err := handler(w, cfg.ToContext(r))
		if err == nil {
			return
		}
		if status < 500 {
			slog.Info("request failed", "method", r.Method, "url", r.URL, "http_status", status, "error", err)
		} else {
			slog.Error("error occurred", "method", r.Method, "url", r.URL, "http_status", status, "error", err)
		}
		http.Error(w, http.StatusText(status), status)
	}
}

func main() {
	parseArgs()

	if printVersion {
		fmt.Println(version)
		os.Exit(0)
	}

	logLevel := slog.LevelInfo
	if verbose {
		logLevel = slog.LevelDebug
	}
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: logLevel})))

	// if metadir is not set we store metadata with tarballs
	if metadir == "" {
		metadir = storageDir
	}
	store := storage.NewFileStore(storageDir, metadir)
	var err error
	cfg, err = config.New(store, registry, urltemplate, proxystash, fetchAll)
	if err != nil {
		slog.Error("error creating config, exiting: %w", err)
		os.Exit(1)
	}

	if indexAll {
		os.Exit(reindexAll(store, pkgthreads))
	}
	if indexPkg != "" {
		os.Exit(reindexPackage(store, indexPkg))
	}

	http.HandleFunc("GET /{pkg}", middleware(handle.PackageMetadata))
	http.HandleFunc("GET /{pkg}/-/{tarball}", middleware(handle.Tarball))
	http.HandleFunc("GET /{scope}/{pkg}/-/{tarball}", middleware(handle.Tarball))
	http.HandleFunc("POST /api/index/{registry}/{pkg}", middleware(handle.Index))
	slog.Info("started enpeeem", "addr", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		slog.Error("server error", "cause", err)
	}
}
