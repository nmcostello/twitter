package main

import (
	"fmt"
	"log"
	"net/http"
	"sync/atomic"
)

func main() {
	const (
		filepathRoot     = "."
		port             = "8080"
		appPath          = "/app/"
		healthzPath      = "/healthz"
		metricsPath      = "/metrics"
		resetMetricspath = "/reset"
	)

	cfg := apiConfig{
		fileserverHits: atomic.Int32{},
	}
	mux := http.NewServeMux()
	srv := http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	fs := http.StripPrefix(appPath, http.FileServer(http.Dir(filepathRoot)))

	mux.Handle(appPath, cfg.middlewareMetricsInc(loggingMiddleware(fs)))
	mux.Handle(healthzPath, loggingMiddleware(http.HandlerFunc(handlerHealthz)))
	mux.Handle(metricsPath, loggingMiddleware(http.HandlerFunc(cfg.handlerHits)))
	mux.Handle(resetMetricspath, loggingMiddleware(http.HandlerFunc(cfg.handlerReset)))

	log.Printf("Starting server at port %s", port)
	log.Fatal(srv.ListenAndServe())
}

type apiConfig struct {
	fileserverHits atomic.Int32
}

func (cfg *apiConfig) handlerHits(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf("Hits: %d", cfg.fileserverHits.Load())))
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

func (cfg *apiConfig) handlerReset(w http.ResponseWriter, req *http.Request) {
	cfg.fileserverHits.Store(0)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Hits reset to 0"))
}

func handlerHealthz(w http.ResponseWriter, req *http.Request) {
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(http.StatusText(http.StatusOK)))
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Received request: method=%s, url=%s, remote_addr=%s", r.Method, r.URL, r.RemoteAddr)
		next.ServeHTTP(w, r) // Call the next handler
	})
}
