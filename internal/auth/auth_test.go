package auth

import (
	"encoding/hex"
	"net/http"
	"testing"

	"github.com/google/uuid"
)

func TestHashPassword(t *testing.T) {
	password := "testpassword"
	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword failed: %v", err)
	}
	if hash == "" {
		t.Error("HashPassword returned empty hash")
	}
	if hash == password {
		t.Error("Hash should not equal password")
	}
}

func TestCheckPasswordHash(t *testing.T) {
	password := "testpassword"
	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword failed: %v", err)
	}

	// Test correct password
	match, err := CheckPasswordHash(password, hash)
	if err != nil {
		t.Fatalf("CheckPasswordHash failed: %v", err)
	}
	if !match {
		t.Error("CheckPasswordHash should return true for correct password")
	}

	// Test incorrect password
	match, err = CheckPasswordHash("wrongpassword", hash)
	if err != nil {
		t.Fatalf("CheckPasswordHash failed: %v", err)
	}
	if match {
		t.Error("CheckPasswordHash should return false for incorrect password")
	}
}

func TestMakeJwt(t *testing.T) {
	userID := uuid.New()
	tokenSecret := "secret"
	expiresIn := int64(3600)

	token, err := MakeJwt(userID, tokenSecret, expiresIn)
	if err != nil {
		t.Fatalf("MakeJwt failed: %v", err)
	}
	if token == "" {
		t.Error("MakeJwt returned empty token")
	}
}

func TestValidateJWT(t *testing.T) {
	userID := uuid.New()
	tokenSecret := "secret"
	expiresIn := int64(3600)

	token, err := MakeJwt(userID, tokenSecret, expiresIn)
	if err != nil {
		t.Fatalf("MakeJwt failed: %v", err)
	}

	validatedUserID, err := ValidateJWT(token, tokenSecret)
	if err != nil {
		t.Fatalf("ValidateJWT failed: %v", err)
	}
	if validatedUserID != userID {
		t.Errorf("ValidateJWT returned wrong userID: got %v, want %v", validatedUserID, userID)
	}

	// Test with wrong secret
	_, err = ValidateJWT(token, "wrongsecret")
	if err == nil {
		t.Error("ValidateJWT should fail with wrong secret")
	}

	// Test with invalid token
	_, err = ValidateJWT("invalidtoken", tokenSecret)
	if err == nil {
		t.Error("ValidateJWT should fail with invalid token")
	}

	// Test with expired token
	expiredToken, err := MakeJwt(userID, tokenSecret, -3600) // expired 1 hour ago
	if err != nil {
		t.Fatalf("MakeJwt failed for expired token: %v", err)
	}
	_, err = ValidateJWT(expiredToken, tokenSecret)
	if err == nil {
		t.Error("ValidateJWT should fail with expired token")
	}
}

func TestGetBearerToken(t *testing.T) {
	// Test valid Bearer token
	headers := http.Header{}
	headers.Set("Authorization", "Bearer validtoken")
	token, err := GetBearerToken(headers)
	if err != nil {
		t.Fatalf("GetBearerToken failed: %v", err)
	}
	if token != "validtoken" {
		t.Errorf("Expected 'validtoken', got '%s'", token)
	}

	// Test missing Authorization header
	headers = http.Header{}
	_, err = GetBearerToken(headers)
	if err == nil {
		t.Error("GetBearerToken should fail with missing header")
	}

	// Test header too short
	headers.Set("Authorization", "Bear")
	_, err = GetBearerToken(headers)
	if err == nil {
		t.Error("GetBearerToken should fail with short header")
	}

	// Test invalid prefix
	headers.Set("Authorization", "Basic token")
	_, err = GetBearerToken(headers)
	if err == nil {
		t.Error("GetBearerToken should fail with invalid prefix")
	}
}

func TestMakeRefreshToken(t *testing.T) {
	token := MakeRefreshToken()
	if token == "" {
		t.Error("MakeRefreshToken returned empty token")
	}
	if len(token) != 64 {
		t.Errorf("MakeRefreshToken returned token of length %d, expected 64", len(token))
	}
	// Check if it's valid hex
	_, err := hex.DecodeString(token)
	if err != nil {
		t.Errorf("MakeRefreshToken returned invalid hex string: %v", err)
	}
}
