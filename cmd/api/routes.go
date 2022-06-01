package api

import (
	"encoding/json"
	"github.com/go-chi/chi"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
)

const (
	contentType     = "Content-Type"
	applicationJSON = "application/json"
)

//Response from the servid
type Response struct {
	Message    string `json:"message"`
	StatusText string `json:"statusText"`
	StatusCode int    `json:"statusCode"`
}

func contentTypeJSON(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set(contentType, applicationJSON)
		h(w, r)
	}
}

// Router structs represents Handlers
type Router struct {
	mux     *chi.Mux
	metrics *prometheus.Registry
}

// NewRouter is creating New Router with Handlers
func NewRouter(mux *chi.Mux, req *prometheus.Registry) *Router {
	return &Router{mux: mux, metrics: req}
}

// Routes , all HTTP routes
func (r *Router) Routes() {
	r.mux.Get("/health", contentTypeJSON(r.health()))
	r.mux.Handle("/metrics", r.prometheus())
}

func (r *Router) prometheus() http.Handler {

	r.metrics.MustRegister(collectors.NewBuildInfoCollector())
	r.metrics.MustRegister(collectors.NewGoCollector(
		collectors.WithGoCollections(collectors.GoRuntimeMetricsCollection),
	))
	// Expose the registered metrics via HTTP.
	return promhttp.HandlerFor(
		r.metrics,
		promhttp.HandlerOpts{
			// Opt into OpenMetrics to support exemplars.
			EnableOpenMetrics: true,
		},
	)
}

func (r *Router) health() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		response := Response{
			Message:    "healthy",
			StatusText: http.StatusText(http.StatusOK),
			StatusCode: http.StatusOK,
		}
		if err := json.NewEncoder(w).Encode(response); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

}
