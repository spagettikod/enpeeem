package config

import (
	"context"
	"enpeeem/storage"
	"net/http"
)

type Config struct {
	Registry   string
	Store      storage.Store
	ProxyStash bool
	FetchAll   bool
}

type cfgKey string

const key cfgKey = "config"

func (cfg Config) ToContext(r *http.Request) *http.Request {
	return r.WithContext(context.WithValue(r.Context(), key, cfg))
}

func FromContext(r *http.Request) Config {
	cfg, _ := r.Context().Value(key).(Config)
	return cfg
}
