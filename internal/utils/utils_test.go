package utils

import (
	"encoding/json"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/SnakeyNinjaGOAT/chirpy/internal/database"
	"github.com/google/uuid"
)

func TestValidateChirp(t *testing.T) {
	// Test valid chirp
	valid, cleaned, length := ValidateChirp("Hello world")
	if !valid {
		t.Error("ValidateChirp should return true for valid chirp")
	}
	if cleaned != "Hello world" {
		t.Errorf("Expected 'Hello world', got '%s'", cleaned)
	}
	if length != 11 {
		t.Errorf("Expected length 11, got %d", length)
	}

	// Test chirp with profanity
	valid, cleaned, length = ValidateChirp("This is kerfuffle")
	if !valid {
		t.Error("ValidateChirp should return true for chirp with profanity")
	}
	if cleaned != "This is ****" {
		t.Errorf("Expected 'This is ****', got '%s'", cleaned)
	}
	if length != 17 {
		t.Errorf("Expected length 17, got %d", length)
	}

	// Test too long chirp
	longChirp := string(make([]byte, 141))
	valid, _, length = ValidateChirp(longChirp)
	if valid {
		t.Error("ValidateChirp should return false for too long chirp")
	}
	if length != 141 {
		t.Errorf("Expected length 141, got %d", length)
	}
}

func TestRespondWithError(t *testing.T) {
	w := httptest.NewRecorder()
	RespondWithError(w, 400, "test error")

	if w.Code != 400 {
		t.Errorf("Expected status 400, got %d", w.Code)
	}

	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Error unmarshaling response: %v", err)
	}

	if response["error"] != "test error" {
		t.Errorf("Expected error 'test error', got '%s'", response["error"])
	}
}

func TestRespondWithJSON(t *testing.T) {
	w := httptest.NewRecorder()
	payload := map[string]string{"key": "value"}
	RespondWithJSON(w, 200, payload)

	if w.Code != 200 {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	if w.Header().Get("Content-Type") != "application/json" {
		t.Errorf("Expected Content-Type application/json, got %s", w.Header().Get("Content-Type"))
	}

	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Error unmarshaling response: %v", err)
	}

	if response["key"] != "value" {
		t.Errorf("Expected 'value', got '%s'", response["key"])
	}
}

func TestConvertUser(t *testing.T) {
	dbUser := database.User{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Email:     "test@example.com",
	}

	user := ConvertUser(dbUser)

	if user.ID != dbUser.ID {
		t.Errorf("Expected ID %v, got %v", dbUser.ID, user.ID)
	}
	if user.Email != dbUser.Email {
		t.Errorf("Expected email %s, got %s", dbUser.Email, user.Email)
	}
}

func TestConvertChirp(t *testing.T) {
	dbChirp := database.Chirp{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Body:      "test chirp",
		UserID:    uuid.New(),
	}

	chirp := ConvertChirp(dbChirp)

	if chirp.ID != dbChirp.ID {
		t.Errorf("Expected ID %v, got %v", dbChirp.ID, chirp.ID)
	}
	if chirp.Body != dbChirp.Body {
		t.Errorf("Expected body %s, got %s", dbChirp.Body, chirp.Body)
	}
	if chirp.UserId != dbChirp.UserID {
		t.Errorf("Expected UserID %v, got %v", dbChirp.UserID, chirp.UserId)
	}
}

func TestGenerateTokens(t *testing.T) {
	userID := uuid.New()
	secret := "testsecret"

	access, refresh, err := GenerateTokens(userID, secret)
	if err != nil {
		t.Fatalf("GenerateTokens failed: %v", err)
	}
	if access == "" {
		t.Error("Access token should not be empty")
	}
	if refresh == "" {
		t.Error("Refresh token should not be empty")
	}
}

func TestGenerateAccessToken(t *testing.T) {
	userID := uuid.New()
	secret := "testsecret"

	token, err := GenerateAccessToken(userID, secret)
	if err != nil {
		t.Fatalf("GenerateAccessToken failed: %v", err)
	}
	if token == "" {
		t.Error("Token should not be empty")
	}
}

func TestGenerateRefreshToken(t *testing.T) {
	token := GenerateRefreshToken()
	if token == "" {
		t.Error("Refresh token should not be empty")
	}
	if len(token) != 64 {
		t.Errorf("Expected length 64, got %d", len(token))
	}
}
