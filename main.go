package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/SnakeyNinjaGOAT/chirpy/internal/config"
	"github.com/SnakeyNinjaGOAT/chirpy/internal/database"
	"github.com/SnakeyNinjaGOAT/chirpy/internal/handlers"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

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

	apiCfg := config.ApiConfig{}
	apiCfg.Db = dbQueries

	apiCfg.JwtSecret = os.Getenv("JWT_TOKEN")

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

func setUpEndPoints(mux *http.ServeMux, apiCfg *config.ApiConfig) {
	mux.Handle("/app/", apiCfg.MiddlewareIncrementHits(handlers.HandleFileServer()))
	mux.HandleFunc("GET /api/healthz", handlers.HandleReadyPath)
	mux.HandleFunc("GET /admin/metrics", handlers.HandleTotalRequests(apiCfg))
	mux.HandleFunc("POST /admin/reset", handlers.HandleReset(apiCfg))

	mux.HandleFunc("POST /api/refresh", handlers.HandleRefresh(apiCfg))
	mux.HandleFunc("POST /api/revoke", handlers.HandleRevoke(apiCfg))

	mux.HandleFunc("GET /api/chirps", handlers.HandleGetChirps(apiCfg))
	mux.HandleFunc("GET /api/chirps/{chirpId}", handlers.HandleGetChirp(apiCfg))
	mux.HandleFunc("POST /api/users", handlers.HandlePostUsers(apiCfg))
	mux.HandleFunc("POST /api/chirps", handlers.HandlePostChirp(apiCfg))
	mux.HandleFunc("POST /api/login", handlers.HandleLogin(apiCfg))

}
