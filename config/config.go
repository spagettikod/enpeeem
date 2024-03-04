package config

import (
	"context"
	"enpeeem/storage"
	"net/http"
	"text/template"
)

type Config struct {
	Registry    string
	Store       storage.Store
	ProxyStash  bool
	FetchAll    bool
	URLTemplate *template.Template
}

type cfgKey string

const key cfgKey = "config"

// default template https://{{.Package.Registry}}/{{.Package.Name}}/{{.Package.Scope}}/-/{{.Name}}
func New(store storage.Store, registry, urltemplate string, proxystash, fetchall bool) (Config, error) {
	cfg := Config{
		Registry:   registry,
		Store:      store,
		ProxyStash: proxystash,
		FetchAll:   fetchall,
	}
	if urltemplate != "" {
		tmpl, err := template.New("rewrite").Parse(urltemplate)
		if err != nil {
			return cfg, err
		}
		cfg.URLTemplate = tmpl
	}
	return cfg, nil
}

func (cfg Config) ToContext(r *http.Request) *http.Request {
	return r.WithContext(context.WithValue(r.Context(), key, cfg))
}

func FromContext(r *http.Request) Config {
	cfg, _ := r.Context().Value(key).(Config)
	return cfg
}
