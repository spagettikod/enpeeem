package main

import (
	"enpeeem/storage"
	"log/slog"
	"sync"

	"github.com/schollz/progressbar/v3"
)

func reindexAll() int {
	pkgs, err := store.Packages()
	if err != nil {
		slog.Error("failed to list packages", "cause", err)
	}
	var bar *progressbar.ProgressBar
	wg := sync.WaitGroup{}
	if progress {
		bar = progressbar.Default(int64(len(pkgs)), "indexing packages")
	}
	exitCode := 0
	for _, pkg := range pkgs {
		wg.Add(1)
		go func() {
			if indexPackage(pkg) != 0 {
				exitCode = 1
			}
			wg.Done()
			if bar != nil {
				bar.Add(1)
			}
		}()
	}
	wg.Wait()
	return exitCode
}

func reindexPackage(pkguri string) int {
	pkg, err := storage.PackageMetadataFromURI(pkguri)
	if err != nil {
		slog.Error("failed to parse package uri", "cause", err)
		return 1
	}
	return indexPackage(pkg)
}

func indexPackage(pkg storage.Package) int {
	if _, err := store.Index(pkg); err != nil {
		slog.Error("error indexing package", "cause", err, "package", pkg.String())
		return 1
	}
	return 0
}
