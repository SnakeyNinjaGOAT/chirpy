package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"sync/atomic"
	"time"

	"github.com/SnakeyNinjaGOAT/chirpy/internal/database"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type User struct {
	Id        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `json:"email"`
}

type Chirp struct {
	Id        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Body      string    `json:"body"`
	UserId    uuid.UUID `json:"user_id"`
}

type apiConfig struct {
	fileserverHits atomic.Int32
	db             *database.Queries
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {

		cfg.fileserverHits.Add(1)
		next.ServeHTTP(writer, request)
	})
}

func main() {
	err := godotenv.Load()
	if err != nil {
		fmt.Println(err)
	}
	dbUrl := os.Getenv("DB_URL")
	db, err := sql.Open("postgres", dbUrl)

	dbQueries := database.New(db)

	if err != nil {
		fmt.Println(err)
	}
	mux := http.NewServeMux()

	apiCfg := apiConfig{}
	apiCfg.db = dbQueries

	if err != nil {
		log.Printf("Error connecting to database: %s", err)
	}

	setUpEndPoints(mux, &apiCfg)

	server := &http.Server{
		Handler: mux,
		Addr:    ":8080",
	}

	err = server.ListenAndServe()

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func setUpEndPoints(mux *http.ServeMux, apiCfg *apiConfig) {
	mux.Handle("/app/", apiCfg.middlewareMetricsInc(handleFileServer()))
	mux.HandleFunc("GET /api/healthz", handleReadyPath)
	mux.HandleFunc("GET /admin/metrics", apiCfg.handlerTotalRequests)
	mux.HandleFunc("POST /admin/reset", apiCfg.handleReset)

	mux.HandleFunc("GET /api/chirps", apiCfg.handleGetChirps)
	mux.HandleFunc("GET /api/chirps/{chirpId}", apiCfg.handleGetChirp)
	mux.HandleFunc("POST /api/users", apiCfg.handlePostUsers)
	mux.HandleFunc("POST /api/chirps", apiCfg.handlePostChirp)

}

func (cfg *apiConfig) handleGetChirp(writer http.ResponseWriter, request *http.Request) {
	chirpId, err := uuid.Parse(request.PathValue("chirpId"))

	if err != nil {
		errMsg := fmt.Sprintf("Error getting chirp: %s", err)
		respondWithError(writer, 400, errMsg)
		return
	}

	chirp, err := cfg.db.GetChirp(request.Context(), chirpId)

	if err != nil {
		errMsg := fmt.Sprintf("Chirp doesn't exist: %s", err)
		respondWithError(writer, 404, errMsg)
		return
	}

	returnChirp := convertChirp(chirp)

	respondWithJSON(writer, 200, returnChirp)
}

func (cfg *apiConfig) handleGetChirps(writer http.ResponseWriter, request *http.Request) {

	chirps, err := cfg.db.GetChirps(request.Context())
	if err != nil {
		errMsg := fmt.Sprintf("Error retrieving all chirps: %s", err)
		respondWithError(writer, 500, errMsg)
		return
	}
	chirpList := make([]Chirp, len(chirps))
	for i, chirp := range chirps {
		chirpList[i] = convertChirp(chirp)
	}

	respondWithJSON(writer, 200, chirpList)

}

func convertChirp(chirp database.Chirp) Chirp {
	return Chirp{
		Id:        chirp.ID,
		CreatedAt: chirp.CreatedAt,
		UpdatedAt: chirp.UpdatedAt,
		Body:      chirp.Body,
		UserId:    chirp.UserID,
	}
}

func (cfg *apiConfig) handlePostChirp(writer http.ResponseWriter, request *http.Request) {
	type chirpRequest struct {
		Body   string    `json:"body"`
		UserId uuid.UUID `json:"user_id"`
	}

	decoder := json.NewDecoder(request.Body)
	var chirp chirpRequest
	err := decoder.Decode(&chirp)

	if err != nil {
		errMsg := fmt.Sprintf("Error decoding chirp: %s", err)
		respondWithError(writer, 500, errMsg)
		return
	}

	validLength, finalBody, chirpLength := validateChirp(chirp.Body)

	if !validLength {
		errMsg := fmt.Sprintf("Chirp exceeds maximium length of (140) with %d characters", chirpLength)
		respondWithError(writer, 400, errMsg)
		return
	}

	_, err = cfg.db.GetUser(request.Context(), chirp.UserId)

	if err != nil {
		errMsg := fmt.Sprintf("User does not exist: %s", err)
		respondWithError(writer, 400, errMsg)
		return
	}

	chirpParams := database.CreateChirpParams{
		Body:   finalBody,
		UserID: chirp.UserId,
	}
	createdChirp, err := cfg.db.CreateChirp(request.Context(), chirpParams)

	if err != nil {
		errMsg := fmt.Sprintf("Couldn't create chirp: %s", err)
		respondWithError(writer, 500, errMsg)
		return
	}

	returnChirp := convertChirp(createdChirp)

	respondWithJSON(writer, 201, returnChirp)

}

func (cfg *apiConfig) handlePostUsers(writer http.ResponseWriter, request *http.Request) {
	type parameters struct {
		Email string `json:"email"`
	}

	decode := json.NewDecoder(request.Body)
	var params parameters
	err := decode.Decode(&params)

	if err != nil {
		errMsg := fmt.Sprintf("Error decoding body: %s", err)
		respondWithError(writer, 500, errMsg)
		return
	}

	user, err := cfg.db.CreateUser(request.Context(), params.Email)

	if err != nil {
		errMsg := fmt.Sprintf("Error creating user: %s", err)
		respondWithError(writer, 500, errMsg)
		return
	}

	data := User{
		Id:        user.ID,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Email:     user.Email,
	}

	respondWithJSON(writer, 201, data)

}

func validateChirp(chirp string) (bool, string, int) {
	chirpLen := len(chirp)
	if chirpLen > 140 {
		return false, chirp, chirpLen
	}

	finalString, _ := removeProfanity(chirp)
	return true, finalString, chirpLen
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

func (cfg *apiConfig) handleReset(writer http.ResponseWriter, request *http.Request) {

	platform := os.Getenv("PLATFORM")

	if platform != "dev" {
		errMsg := "Forbidden"
		respondWithError(writer, 403, errMsg)
		return
	}

	err := cfg.db.DeleteAllUsers(request.Context())

	if err != nil {
		errMsg := fmt.Sprintf("Error deleting from users table: %s", err)
		respondWithError(writer, 500, errMsg)
		return
	}

	// Reset metrics
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
