package utils

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"unicode/utf8"

	"github.com/SnakeyNinjaGOAT/chirpy/internal/auth"
	"github.com/SnakeyNinjaGOAT/chirpy/internal/database"
	"github.com/SnakeyNinjaGOAT/chirpy/internal/models"
	"github.com/google/uuid"
)

func ValidateChirp(chirp string) (bool, string, int) {
	chirpLen := utf8.RuneCountInString(chirp)

	if chirpLen > 140 {
		return false, chirp, chirpLen
	}

	finalString := removeProfanity(chirp)
	return true, finalString, chirpLen
}

func removeProfanity(chirp string) string {
	words := strings.Split(chirp, " ")

	for i, word := range words {
		if _, found := profaneWords[strings.ToLower(word)]; found {
			words[i] = "****"
		}
	}

	return strings.Join(words, " ")

}

var profaneWords = map[string]bool{
	"kerfuffle": true,
	"sharbert":  true,
	"fornax":    true,
}

func RespondWithError(writer http.ResponseWriter, code int, message string) {
	type errorJson struct {
		Error string `json:"error"`
	}

	errValue := errorJson{
		Error: message,
	}

	writer.WriteHeader(code)
	data, err := json.Marshal(errValue)

	if err != nil {
		log.Printf("error marshaling error response: %s", err)
	}

	writer.Write(data)
	log.Print(message)
}

func RespondWithJSON(writer http.ResponseWriter, code int, payload interface{}) {
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(code)
	data, err := json.Marshal(payload)

	if err != nil {
		errMsg := fmt.Sprintf("Error marshaling payload: %s", err)
		RespondWithError(writer, 500, errMsg)
	}

	writer.Write(data)
}

func ConvertUser(user database.User) models.User {
	return models.User{
		ID:          user.ID,
		CreatedAt:   user.CreatedAt,
		UpdatedAt:   user.UpdatedAt,
		Email:       user.Email,
		IsChirpyRed: user.IsChirpyRed,
	}
}

func ConvertChirp(chirp database.Chirp) models.Chirp {
	return models.Chirp{
		ID:        chirp.ID,
		CreatedAt: chirp.CreatedAt,
		UpdatedAt: chirp.UpdatedAt,
		Body:      chirp.Body,
		UserId:    chirp.UserID,
	}
}

func GenerateTokens(userID uuid.UUID, secret string) (accessToken, refreshToken string, err error) {
	accessToken, err = GenerateAccessToken(userID, secret)

	if err != nil {
		return
	}

	refreshToken = GenerateRefreshToken()
	return

}

func GenerateAccessToken(userID uuid.UUID, secret string) (string, error) {
	return auth.MakeJwt(userID, secret, 3600)
}

func GenerateRefreshToken() string {
	return auth.MakeRefreshToken()
}
