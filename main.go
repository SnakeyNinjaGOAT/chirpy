package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
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

	setUpEndPoints(mux, &apiCfg)

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

func setUpEndPoints(mux *http.ServeMux, apiCfg *apiConfig) {
	mux.Handle("/app/", apiCfg.middlewareMetricsInc(handleFileServer()))
	mux.HandleFunc("GET /api/healthz", handleReadyPath)
	mux.HandleFunc("GET /admin/metrics", apiCfg.handlerTotalRequests)
	mux.HandleFunc("POST /admin/reset", apiCfg.handleMetricReset)
	mux.HandleFunc("POST /api/validate_chirp", handleValidateChirp)

}

func handleValidateChirp(writer http.ResponseWriter, request *http.Request) {
	type parameters struct {
		Body string `json:"body"`
	}
	decoder := json.NewDecoder(request.Body)

	params := parameters{}

	err := decoder.Decode(&params)

	if err != nil {
		errMsg := fmt.Sprintf("Error decoding parameters: %s", err)
		respondWithError(writer, 500, errMsg)
		return
	}

	type returnVal struct {
		CleanedBody string `json:"cleaned_body"`
	}

	if len(params.Body) > 140 {
		respondWithError(writer, 400, "This chirp is too long (max 140 characters)")
		return
	}

	data, _ := removeProfanity(params.Body)
	payload := returnVal{}

	payload.CleanedBody = data

	respondWithJSON(writer, 200, payload)

}

func respondWithError(writer http.ResponseWriter, code int, msg string) {
	type errorJson struct {
		Error string `json:"error"`
	}
	errValue := errorJson{
		Error: msg,
	}
	writer.WriteHeader(code)
	data, err := json.Marshal(errValue)

	if err != nil {
		log.Printf("Error marsheling error response: %s", err)
		return
	}

	writer.Write([]byte(data))
	log.Print(msg)
}

func removeProfanity(chirp string) (string, bool) {
	words := strings.Split(chirp, " ")

	profaneWords := map[string]bool{}
	profaneWords["kerfuffle"] = true
	profaneWords["sharbert"] = true
	profaneWords["fornax"] = true

	containsProfanity := false
	for i, word := range words {
		if _, found := profaneWords[strings.ToLower(word)]; found {
			words[i] = "****"
			containsProfanity = true
		}
	}

	return strings.Join(words, " "), containsProfanity
}

func respondWithJSON(writer http.ResponseWriter, code int, payload interface{}) {
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(code)
	data, err := json.Marshal(payload)

	if err != nil {
		errMsg := fmt.Sprintf("Error marsheling payload: %s", err)
		respondWithError(writer, 500, errMsg)
	}

	writer.Write(data)
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
	writer.Header().Set("Content-Type", "text/html; charset=utf-8")
	body := fmt.Sprintf(`<html>

<body>
    <h1>Welcome, Chirpy Admin</h1>
    <p>Chirpy has been visited %d times!</p>
</body>

</html>`, cfg.fileserverHits.Load())
	writer.WriteHeader(http.StatusOK)

	writer.Write([]byte(body))

}
