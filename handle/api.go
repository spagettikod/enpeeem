package handle

import (
	"enpeeem/config"
	"enpeeem/storage"
	"log/slog"
	"net/http"
)

var processing = map[string]bool{}

func Index(w http.ResponseWriter, r *http.Request) {
	cfg := config.FromContext(r)
	s, p := splitPkg(r.PathValue("pkg"))
	pkg, err := storage.NewPackage(r.PathValue("registry"), s, p)
	if err != nil {
		logErr(w, r, http.StatusInternalServerError, err)
		return
	}
	if _, found := processing[pkg.String()]; !found {
		slog.Debug("start indexing package", "pkg", pkg.String())
		processing[pkg.String()] = true
		defer delete(processing, pkg.String())
		if _, err = cfg.Store.Index(pkg); err != nil {
			logErr(w, r, http.StatusInternalServerError, err)
			return
		}
		return
	}
	slog.Debug("package already being indexed", "pkg", pkg.String())
	// returns HTTP 102 if package is already being indexed
	http.Error(w, http.StatusText(http.StatusProcessing), http.StatusProcessing)
}
