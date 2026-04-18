package main

import (
	"fmt"
	"net/http"
	"os"
	"sync/atomic"
)

type apiConfig struct {
	fileserverHits atomic.Int32
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {

		cfg.fileserverHits.Add(1)
		next.ServeHTTP(writer, request)
	})
}

func main() {
	mux := http.NewServeMux()

	apiCfg := apiConfig{}

	mux.Handle("/app/", apiCfg.middlewareMetricsInc(handleFileServer()))
	mux.HandleFunc("GET /api/healthz", handleReadyPath)
	mux.HandleFunc("GET /api/metrics", apiCfg.handlerTotalRequests)
	mux.HandleFunc("POST /api/reset", apiCfg.handleMetricReset)

	server := &http.Server{
		Handler: mux,
		Addr:    ":8080",
	}

	err := server.ListenAndServe()

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func (cfg *apiConfig) handleMetricReset(writer http.ResponseWriter, request *http.Request) {
	oldHits := cfg.fileserverHits.Swap(0)
	msg := fmt.Sprintf("Metrics Reset\nPrevious hits: %v", oldHits)
	writer.Header().Set("Content-Type", "text/plain; charset=utf-8")
	writer.WriteHeader(http.StatusOK)
	writer.Write([]byte(msg))

}

func handleFileServer() http.Handler {
	return http.StripPrefix("/app", http.FileServer(http.Dir("./static/")))
}

func handleReadyPath(writer http.ResponseWriter, request *http.Request) {
	writer.Header().Set("Content-Type", "text/plain; charset=utf-8")
	writer.WriteHeader(http.StatusOK)
	writer.Write([]byte("OK"))

}

func (cfg *apiConfig) handlerTotalRequests(writer http.ResponseWriter, request *http.Request) {
	writer.Header().Set("Content-Type", "text/plain; charset=utf-8")

	writer.WriteHeader(http.StatusOK)
	totalHits := fmt.Sprintf("Hits: %v", cfg.fileserverHits.Load())

	writer.Write([]byte(totalHits))

}
