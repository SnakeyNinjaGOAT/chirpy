package config

import (
	"context"
	"database/sql"
	"net/http"
	"os"
	"sync/atomic"

	"github.com/SnakeyNinjaGOAT/chirpy/internal/database"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
)

type DBInterface interface {
	GetRefreshToken(context.Context, string) (database.RefreshToken, error)
	RevokeRefreshToken(context.Context, string) error
	GetUserByEmail(context.Context, string) (database.User, error)
	CreateUser(context.Context, database.CreateUserParams) (database.User, error)
	GetUser(context.Context, uuid.UUID) (database.User, error)
	CreateChirp(context.Context, database.CreateChirpParams) (database.Chirp, error)
	GetChirp(context.Context, uuid.UUID) (database.Chirp, error)
	GetChirps(context.Context, bool) ([]database.Chirp, error)
	CreateRefreshToken(context.Context, database.CreateRefreshTokenParams) (database.RefreshToken, error)
	DeleteAllUsers(context.Context) error
	UpdateUser(context.Context, database.UpdateUserParams) (database.User, error)
	DeleteChirp(context.Context, uuid.UUID) error
	UpgradeUserToChirpyRed(context.Context, uuid.UUID) error
	GetChirpsByUser(context.Context, database.GetChirpsByUserParams) ([]database.Chirp, error)
}

type ApiConfig struct {
	FileserverHits atomic.Int32
	Db             DBInterface
	JwtSecret      string
	PolkaKey       string
}

func LoadEnv() (ApiConfig, error) {
	err := godotenv.Load()

	if err != nil {
		return ApiConfig{}, err
	}

	dbUrl := os.Getenv("DB_URL")
	jwtSecret := os.Getenv("JWT_SECRET")
	polkaKey := os.Getenv("POLKA_KEY")
	db, err := sql.Open("postgres", dbUrl)

	if err != nil {
		return ApiConfig{}, err
	}

	dbQueries := database.New(db)

	return ApiConfig{
		Db:        dbQueries,
		JwtSecret: jwtSecret,
		PolkaKey:  polkaKey,
	}, nil

}

func (cfg *ApiConfig) MiddlewareIncrementHits(next http.Handler) http.HandlerFunc {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		cfg.FileserverHits.Add(1)
		next.ServeHTTP(writer, request)
	})
}
