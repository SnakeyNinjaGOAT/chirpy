package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/SnakeyNinjaGOAT/chirpy/internal/auth"
	config "github.com/SnakeyNinjaGOAT/chirpy/internal/config"
	"github.com/SnakeyNinjaGOAT/chirpy/internal/database"
	"github.com/SnakeyNinjaGOAT/chirpy/internal/models"
	"github.com/SnakeyNinjaGOAT/chirpy/internal/utils"
	"github.com/google/uuid"
)

func HandleRevoke(cfg *config.ApiConfig) http.HandlerFunc {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {

		refreshToken, err := auth.GetBearerToken(request.Header)

		if err != nil {
			errMsg := fmt.Sprintf("Error Retrieving Bearer Token: %s", err)
			utils.RespondWithError(writer, 400, errMsg)
			return
		}

		databaseToken, err := cfg.Db.GetRefreshToken(request.Context(), refreshToken)

		if err != nil || databaseToken.RevokedAt.Valid {
			errMsg := "Invalid Token Provided"
			utils.RespondWithError(writer, 401, errMsg)
			return
		}

		err = cfg.Db.RevokeRefreshToken(request.Context(), databaseToken.Token)
		if err != nil {
			errMsg := fmt.Sprintf("error Revoking Token: %s", err)
			utils.RespondWithError(writer, 500, errMsg)
			return
		}

		writer.WriteHeader(http.StatusNoContent)
	})
}

func HandleRefresh(cfg *config.ApiConfig) http.HandlerFunc {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		oldRefreshToken, err := auth.GetBearerToken(request.Header)

		if err != nil {
			errMsg := fmt.Sprintf("Error getting Bearer token: %s", err)
			utils.RespondWithError(writer, 400, errMsg)
			return
		}

		databaseToken, err := cfg.Db.GetRefreshToken(request.Context(), oldRefreshToken)

		if err != nil {
			errMsg := fmt.Sprintf("Error getting Refresh token: %s", err)
			utils.RespondWithError(writer, 401, errMsg)
			return
		} else if databaseToken.RevokedAt.Valid {
			errMsg := fmt.Sprintf("Error, Refresh Token Revoked At: %v", databaseToken.RevokedAt.Time)
			utils.RespondWithError(writer, 401, errMsg)
			return
		}

		if time.Now().After(databaseToken.ExpiresAt) {
			errMsg := fmt.Sprintf("Error, Refresh Token Expired: %v", databaseToken.ExpiresAt)
			utils.RespondWithError(writer, 401, errMsg)
			return

		}

		newToken, err := utils.GenerateAccessToken(databaseToken.UserID, cfg.JwtSecret)

		if err != nil {
			errMsg := fmt.Sprintf("Error Generating Access Token: %s", err)
			utils.RespondWithError(writer, 500, errMsg)
			return
		}

		type tokenResponse struct {
			Token string `json:"token"`
		}
		token := tokenResponse{
			Token: newToken,
		}
		utils.RespondWithJSON(writer, 200, token)
	})
}

func HandleLogin(cfg *config.ApiConfig) http.HandlerFunc {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		decoder := json.NewDecoder(request.Body)
		var params models.UserParams
		err := decoder.Decode(&params)

		if err != nil {
			errMsg := fmt.Sprintf("Error decoding body: %s", err)
			utils.RespondWithError(writer, 500, errMsg)
			return
		}

		targetUser, err := cfg.Db.GetUserByEmail(request.Context(), params.Email)

		if err != nil {
			errMsg := fmt.Sprintf("User does not exist: %s", err)
			utils.RespondWithError(writer, 400, errMsg)
			return
		}

		passwordMatch, err := auth.CheckPasswordHash(params.Password, targetUser.HashedPassword)

		if err != nil {
			errMsg := fmt.Sprintf("Could not Check and Verify password: %s", err)
			utils.RespondWithError(writer, 500, errMsg)
			return
		}

		if !passwordMatch {
			errMsg := "Incorrect password"
			utils.RespondWithError(writer, 401, errMsg)
			return
		}

		accessToken, refreshToken, err := utils.GenerateTokens(targetUser.ID, cfg.JwtSecret)

		if err != nil {
			errMsg := fmt.Sprintf("Error generating JWT: %s", err)
			utils.RespondWithError(writer, 500, errMsg)
			return
		}

		refreshTokenParams := database.CreateRefreshTokenParams{
			Token:     refreshToken,
			UserID:    targetUser.ID,
			ExpiresAt: time.Now().Add(60 * 24 * time.Hour),
		}

		_, err = cfg.Db.CreateRefreshToken(request.Context(), refreshTokenParams)

		if err != nil {
			errMsg := fmt.Sprintf("Error creating refresh token: %s", err)
			utils.RespondWithError(writer, 500, errMsg)
			return
		}

		userResponse := utils.ConvertUser(targetUser)
		userResponse.AccessToken = accessToken
		userResponse.RefreshToken = refreshToken

		utils.RespondWithJSON(writer, 200, userResponse)
	})
}

func HandleGetChirp(cfg *config.ApiConfig) http.HandlerFunc {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		chirpId, err := uuid.Parse(request.PathValue("chirpId"))

		if err != nil {
			errMsg := fmt.Sprintf("Error getting chirp: %s", err)
			utils.RespondWithError(writer, 400, errMsg)
			return
		}

		chirp, err := cfg.Db.GetChirp(request.Context(), chirpId)

		if err != nil {
			errMsg := fmt.Sprintf("Chirp doesn't exist: %s", err)
			utils.RespondWithError(writer, 404, errMsg)
			return
		}

		returnChirp := utils.ConvertChirp(chirp)

		utils.RespondWithJSON(writer, 200, returnChirp)
	})
}

func HandleGetChirps(cfg *config.ApiConfig) http.HandlerFunc {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		chirps, err := cfg.Db.GetChirps(request.Context())
		if err != nil {
			errMsg := fmt.Sprintf("Error retrieving all chirps: %s", err)
			utils.RespondWithError(writer, 500, errMsg)
			return
		}
		chirpList := make([]models.Chirp, len(chirps))
		for i, chirp := range chirps {
			chirpList[i] = utils.ConvertChirp(chirp)
		}

		utils.RespondWithJSON(writer, 200, chirpList)
	})
}

func HandlePostChirp(cfg *config.ApiConfig) http.HandlerFunc {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		type chirpRequest struct {
			Body string `json:"body"`
		}

		decoder := json.NewDecoder(request.Body)
		var chirp chirpRequest
		err := decoder.Decode(&chirp)

		if err != nil {
			errMsg := fmt.Sprintf("Error decoding chirp: %s", err)
			utils.RespondWithError(writer, 500, errMsg)
			return
		}

		token, err := auth.GetBearerToken(request.Header)
		if err != nil {
			errMsg := fmt.Sprintf("Error getting Bearer Token: %s", err)
			utils.RespondWithError(writer, 400, errMsg)
			return
		}

		TokenUserId, err := auth.ValidateJWT(token, cfg.JwtSecret)

		if err != nil {
			errMsg := fmt.Sprintf("Error validating jwt token: %s", err)
			utils.RespondWithError(writer, 401, errMsg)
			return
		}

		validLength, finalBody, chirpLength := utils.ValidateChirp(chirp.Body)

		if !validLength {
			errMsg := fmt.Sprintf("Chirp exceeds maximium length of (140) with %d characters", chirpLength)
			utils.RespondWithError(writer, 400, errMsg)
			return
		}

		_, err = cfg.Db.GetUser(request.Context(), TokenUserId)

		if err != nil {
			errMsg := fmt.Sprintf("User does not exist: %s", err)
			utils.RespondWithError(writer, 400, errMsg)
			return
		}

		chirpParams := database.CreateChirpParams{
			Body:   finalBody,
			UserID: TokenUserId,
		}
		createdChirp, err := cfg.Db.CreateChirp(request.Context(), chirpParams)

		if err != nil {
			errMsg := fmt.Sprintf("Couldn't create chirp: %s", err)
			utils.RespondWithError(writer, 500, errMsg)
			return
		}

		returnChirp := utils.ConvertChirp(createdChirp)

		utils.RespondWithJSON(writer, 201, returnChirp)
	})
}

func HandlePostUsers(cfg *config.ApiConfig) http.HandlerFunc {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		decode := json.NewDecoder(request.Body)
		var params models.UserParams
		err := decode.Decode(&params)

		if err != nil {
			errMsg := fmt.Sprintf("Error decoding body: %s", err)
			utils.RespondWithError(writer, 500, errMsg)
			return
		}

		hashedPassword, err := auth.HashPassword(params.Password)
		if err != nil {
			errMsg := fmt.Sprintf("Error hashing password: %s", err)
			utils.RespondWithError(writer, 500, errMsg)
			return
		}

		userParams := database.CreateUserParams{
			Email:          params.Email,
			HashedPassword: hashedPassword,
		}

		user, err := cfg.Db.CreateUser(request.Context(), userParams)

		if err != nil {
			errMsg := fmt.Sprintf("Error creating user: %s", err)
			utils.RespondWithError(writer, 500, errMsg)
			return
		}

		data := utils.ConvertUser(user)

		utils.RespondWithJSON(writer, 201, data)
	})
}

func HandleReset(cfg *config.ApiConfig) http.HandlerFunc {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		platform := os.Getenv("PLATFORM")

		if platform != "dev" {
			errMsg := "Forbidden"
			utils.RespondWithError(writer, 403, errMsg)
			return
		}

		err := cfg.Db.DeleteAllUsers(request.Context())

		if err != nil {
			errMsg := fmt.Sprintf("Error deleting from users table: %s", err)
			utils.RespondWithError(writer, 500, errMsg)
			return
		}

		// Reset metrics
		oldHits := cfg.FileserverHits.Swap(0)
		msg := fmt.Sprintf("Metrics Reset\nPrevious hits: %v", oldHits)
		writer.Header().Set("Content-Type", "text/plain; charset=utf-8")
		writer.WriteHeader(http.StatusOK)
		writer.Write([]byte(msg))
	})
}

func HandleFileServer() http.Handler {
	return http.StripPrefix("/app", http.FileServer(http.Dir("./static/")))
}

func HandleReadyPath(writer http.ResponseWriter, request *http.Request) {
	writer.Header().Set("Content-Type", "text/plain; charset=utf-8")
	writer.WriteHeader(http.StatusOK)
	writer.Write([]byte("OK"))
}

func HandleTotalRequests(cfg *config.ApiConfig) http.HandlerFunc {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("Content-Type", "text/html; charset=utf-8")
		body := fmt.Sprintf(`<html>

	<body>
		<h1>Welcome, Chirpy Admin</h1>
		<p>Chirpy has been visited %d times!</p>
	</body>

	</html>`, cfg.FileserverHits.Load())
		writer.WriteHeader(http.StatusOK)

		writer.Write([]byte(body))
	})
}

func HandleUpdateUser(cfg *config.ApiConfig) http.HandlerFunc {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		accessToken, err := auth.GetBearerToken(request.Header)
		fmt.Println("Access Token: ", accessToken)

		if err != nil {
			errMsg := fmt.Sprintf("Error accessing Bearer token: %s", err)
			utils.RespondWithError(writer, 401, errMsg)
			return
		}

		userID, err := auth.ValidateJWT(accessToken, cfg.JwtSecret)

		if err != nil {
			errMsg := fmt.Sprintf("Error expired or invalid token: %s", err)
			utils.RespondWithError(writer, 401, errMsg)
			return
		}

		userParams := models.UserParams{}
		decoder := json.NewDecoder(request.Body)
		err = decoder.Decode(&userParams)

		if err != nil {
			errMsg := fmt.Sprintf("Error decoding body: %s", err)
			utils.RespondWithError(writer, 400, errMsg)
			return
		}

		newPass, err := auth.HashPassword(userParams.Password)

		if err != nil {
			errMsg := fmt.Sprintf("Error hashing password: %s", err)
			utils.RespondWithError(writer, 500, errMsg)
			return
		}

		updateParams := database.UpdateUserParams{
			ID:             userID,
			Email:          userParams.Email,
			HashedPassword: newPass,
		}

		updatedUser, err := cfg.Db.UpdateUser(request.Context(), updateParams)
		if err != nil {
			errMsg := fmt.Sprintf("Error updating user: %s", err)
			utils.RespondWithError(writer, 500, errMsg)
			return
		}

		utils.RespondWithJSON(writer, 200, utils.ConvertUser(updatedUser))

	})

}

func HandleDeleteChirp(cfg *config.ApiConfig) http.HandlerFunc {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {

		accessToken, err := auth.GetBearerToken(request.Header)

		if err != nil {
			errMsg := fmt.Sprintf("Error invalid access token: %s", err)
			utils.RespondWithError(writer, 401, errMsg)
			return
		}

		userID, err := auth.ValidateJWT(accessToken, cfg.JwtSecret)

		if err != nil {
			errMsg := fmt.Sprintf("Error invalid access token: %s", err)
			utils.RespondWithError(writer, 403, errMsg)
			return
		}
		chirpId, err := uuid.Parse(request.PathValue("chirpId"))

		if err != nil {
			errMsg := fmt.Sprintf("Error invalid chirp: %s", err)
			utils.RespondWithError(writer, 403, errMsg)
			return
		}

		dbChirp, err := cfg.Db.GetChirp(request.Context(), chirpId)

		if dbChirp.UserID != userID {
			errMsg := "Forbidden"
			utils.RespondWithError(writer, 403, errMsg)
			return
		}

		err = cfg.Db.DeleteChirp(request.Context(), dbChirp.ID)
		if err != nil {
			errMsg := fmt.Sprintf("Error deleting chirp: %s", err)
			utils.RespondWithError(writer, 500, errMsg)
			return
		}

		writer.WriteHeader(http.StatusNoContent)

	})
}

func HandleChirpyRedUpgrade(cfg *config.ApiConfig) http.HandlerFunc {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {

		apiKey, err := auth.GetApiKeyToken(request.Header)

		if err != nil {
			utils.RespondWithError(writer, 401, err.Error())
			return
		}

		if apiKey != cfg.PolkaKey {
			errMsg := "Invalid ApiKey"
			utils.RespondWithError(writer, 401, errMsg)
			return
		}

		type data struct {
			UserId uuid.UUID `json:"user_id"`
		}

		type upgradeToRedRequest struct {
			Event string `json:"event"`
			Data  data   `json:"data"`
		}

		var requestData upgradeToRedRequest
		decoder := json.NewDecoder(request.Body)
		err = decoder.Decode(&requestData)

		if err != nil {
			errMsg := fmt.Sprintf("Error decoding request body: %s")
			utils.RespondWithError(writer, 400, errMsg)
			return
		}

		event := strings.Split(requestData.Event, ".")
		target := event[0]
		action := event[1]

		if target != "user" || action != "upgraded" || len(event) != 2 {
			writer.WriteHeader(http.StatusNoContent)
			return
		}

		err = cfg.Db.UpgradeUserToChirpyRed(request.Context(), requestData.Data.UserId)
		if err != nil {
			errMsg := fmt.Sprintf("Error user does not exist: %s", err)
			utils.RespondWithError(writer, 404, errMsg)
			return
		}

		writer.WriteHeader(http.StatusNoContent)

	})
}
