package main

import (
	"enpeeem/storage"
	"log/slog"
	"os"

	"github.com/alitto/pond"
	"github.com/schollz/progressbar/v3"
)

func reindexAll(store storage.Store, pkgthreads int) int {
	pkgs, err := store.Packages()
	if err != nil {
		slog.Error("failed to list packages", "cause", err)
	}
	var bar *progressbar.ProgressBar
	if progress {
		bar = progressbar.NewOptions(len(pkgs), progressbar.OptionSetDescription("indexing packages"), progressbar.OptionSetWriter(os.Stdout), progressbar.OptionShowCount(), progressbar.OptionFullWidth())
	} else {
		bar = progressbar.DefaultSilent(int64(len(pkgs)))
	}
	exitCode := 0
	pool := pond.New(pkgthreads, 0)
	for _, pkg := range pkgs {
		pool.Submit(func() {
			if indexPackage(store, pkg) != 0 {
				exitCode = 1
			}
			bar.Add(1)
		})
	}
	pool.StopAndWait()
	return exitCode
}

func reindexPackage(store storage.Store, pkguri string) int {
	pkg, err := storage.PackageMetadataFromURI(pkguri)
	if err != nil {
		slog.Error("failed to parse package uri", "cause", err)
		return 1
	}
	return indexPackage(store, pkg)
}

func indexPackage(store storage.Store, pkg storage.Package) int {
	if _, err := store.Index(pkg); err != nil {
		slog.Error("error indexing package", "cause", err, "package", pkg.String())
		return 1
	}
	return 0
}
