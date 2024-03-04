package handle

import (
	"enpeeem/config"
	"enpeeem/storage"
	"log/slog"
	"net/http"
)

var processing = map[string]bool{}

func Index(w http.ResponseWriter, r *http.Request) (int, error) {
	cfg := config.FromContext(r)
	isAsync := r.URL.Query().Get("async") == "true"
	s, p := splitPkg(r.PathValue("pkg"))
	pkg, err := storage.NewPackage(r.PathValue("registry"), s, p)
	if err != nil {
		return http.StatusInternalServerError, err
	}
	if _, found := processing[pkg.String()]; !found {
		slog.Debug("start indexing package", "pkg", pkg.String(), "async", isAsync)
		if isAsync {
			go func() {
				if err := doIndex(pkg, cfg); err != nil {
					slog.Error("error occurred", "method", r.Method, "url", r.URL, "cause", err)
					return
				}
			}()
			w.WriteHeader(http.StatusAccepted)
			return http.StatusAccepted, nil
		} else {
			if err := doIndex(pkg, cfg); err != nil {
				return http.StatusInternalServerError, err
			}
			w.WriteHeader(http.StatusNoContent)
			return http.StatusNoContent, nil
		}
	}
	slog.Debug("package already being indexed", "pkg", pkg.String())
	// returns HTTP 429 if package is already being indexed
	return http.StatusTooManyRequests, nil
}

func doIndex(pkg storage.Package, cfg config.Config) error {
	processing[pkg.String()] = true
	defer delete(processing, pkg.String())
	_, err := cfg.Store.Index(pkg)
	return err
}
