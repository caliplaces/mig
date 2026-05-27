// Package httpapi exposes the muscle-group image generator as an HTTP API.
// Routing is built on the Go 1.22+ http.ServeMux pattern matcher; no
// third-party router dependency.
package httpapi

import (
	"net/http"

	"github.com/caliplaces/mig/internal/compositor"
)

// NewRouter returns an http.Handler with all routes and middleware wired up.
func NewRouter(c *compositor.Compositor) http.Handler {
	h := &handlers{c: c}
	mux := http.NewServeMux()

	mux.HandleFunc("GET /{$}", h.root)
	mux.HandleFunc("GET /healthz", h.health)
	mux.HandleFunc("GET /openapi.yaml", h.openapiSpec)
	mux.HandleFunc("GET /docs", h.swaggerUI)
	mux.HandleFunc("GET /getMuscleGroups", h.getMuscleGroups)
	mux.HandleFunc("GET /getBaseImage", h.getBaseImage)
	mux.HandleFunc("GET /getImage", h.getImage)
	mux.HandleFunc("GET /getMulticolorImage", h.getMulticolorImage)
	mux.HandleFunc("GET /getIndividualColorImage", h.getIndividualColorImage)

	return chain(mux, recoverer, logger, cors)
}

// chain wraps h with middlewares. Each middleware listed is applied as an
// outer wrapper, so the first middleware passed runs closest to the
// handler and the last runs first on every request.
func chain(h http.Handler, mws ...func(http.Handler) http.Handler) http.Handler {
	for _, mw := range mws {
		h = mw(h)
	}
	return h
}
