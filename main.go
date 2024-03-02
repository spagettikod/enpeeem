package main

import (
	"enpeeem/storage"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
)

var (
	registry     string
	storageDir   string
	addr         string
	proxystash   bool
	indexAll     bool
	progress     bool
	indexPkg     string
	fetchAll     bool
	store        storage.Store
	version      = "SET VERSION IN MAKEFILE"
	printVersion bool
)

func init() {
	flag.StringVar(&addr, "addr", ":8080", "network address of local registry")
	flag.StringVar(&registry, "registry", "https://registry.npmjs.org", "remote npm registry to use when the flag proxystash is set")
	flag.BoolVar(&indexAll, "index-all", false, "re-index all packages")
	flag.BoolVar(&progress, "progress", false, "show progress where applicable")
	flag.BoolVar(&printVersion, "version", false, "print version")
	flag.BoolVar(&fetchAll, "fetch-all", false, "download all tarbal versions at once if a tarball is not found locally")
	flag.StringVar(&indexPkg, "index", "", "re-index with given package URI, example registry.npmjs.org/@types/react")
	flag.BoolVar(&proxystash, "proxystash", false, "proxy and download to storage if file is not available at storage path")
	flag.Usage = printUsage
}

func printUsage() {
	fmt.Printf(`Local npm registry and proxy.
	
Packages are served from the given path. If the flag proxypath is set the
request will be proxied to the remote registry and the result stored at
the given path.

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

func main() {
	parseArgs()

	if printVersion {
		fmt.Println(version)
		os.Exit(0)
	}

	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelInfo})))

	store = storage.NewFileStore(storageDir)
	if indexAll {
		os.Exit(reindexAll())
	}
	if indexPkg != "" {
		os.Exit(reindexPackage(indexPkg))
	}

	http.HandleFunc("GET /{pkg}", packageMetadataHandler)
	http.HandleFunc("GET /{pkg}/-/{tarball}", tarballHandler)
	http.HandleFunc("GET /{scope}/{pkg}/-/{tarball}", tarballHandler)
	slog.Info("started enpeeem", "addr", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		slog.Error("server error", "cause", err)
	}
}
