package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
)

var (
	registry    string
	storageDir  string
	addr        string
	proxyNStash = false
)

func init() {
	flag.StringVar(&addr, "addr", "localhost:8080", "network address of local registry")
	flag.StringVar(&registry, "registry", "https://registry.npmjs.org", "remote npm registry to use when the flag proxystash is set")
	flag.BoolVar(&proxyNStash, "proxystash", false, "proxy and download to storage if file is not available at storage path")
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
	if len(args) != 1 {
		fmt.Println("error: too few arguments")
		printUsage()
		os.Exit(1)
	}
	storageDir = args[0]
}

func main() {
	parseArgs()
	http.HandleFunc("GET /{pkg}", pkgHandler)
	http.HandleFunc("GET /{pkg}/-/{tarball}", tarballHandler)
	http.HandleFunc("GET /{pkg}/{subpkg}/-/{tarball}", subpackageTarballHandler)
	log.Printf("enpeeem, listening at %s", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalln(err)
	}
}
