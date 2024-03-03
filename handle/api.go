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
	isAsync := r.URL.Query().Get("async") == "true"
	s, p := splitPkg(r.PathValue("pkg"))
	pkg, err := storage.NewPackage(r.PathValue("registry"), s, p)
	if err != nil {
		logErr(w, r, http.StatusInternalServerError, err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	if _, found := processing[pkg.String()]; !found {
		slog.Debug("start indexing package", "pkg", pkg.String(), "async", isAsync)
		if isAsync {
			go func() {
				if err := doIndex(pkg, cfg); err != nil {
					logErr(w, r, http.StatusInternalServerError, err)
					return
				}
			}()
			w.WriteHeader(http.StatusAccepted)
			return
		} else {
			if err := doIndex(pkg, cfg); err != nil {
				logErr(w, r, http.StatusInternalServerError, err)
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusNoContent)
			return
		}
	}
	slog.Debug("package already being indexed", "pkg", pkg.String())
	// returns HTTP 429 if package is already being indexed
	http.Error(w, http.StatusText(http.StatusTooManyRequests), http.StatusTooManyRequests)
}

func doIndex(pkg storage.Package, cfg config.Config) error {
	processing[pkg.String()] = true
	defer delete(processing, pkg.String())
	_, err := cfg.Store.Index(pkg)
	return err
}
