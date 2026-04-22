package auth

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/alexedwards/argon2id"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func HashPassword(password string) (string, error) {
	hashed_password, err := argon2id.CreateHash(password, argon2id.DefaultParams)
	if err != nil {
		return "", err
	}

	return hashed_password, nil
}

func CheckPasswordHash(password string, hash string) (bool, error) {
	isCorrectPassword, err := argon2id.ComparePasswordAndHash(password, hash)
	if err != nil {
		return false, err
	}

	return isCorrectPassword, nil
}

func MakeJwt(userId uuid.UUID, tokenSecret string, expiresIn int64) (string, error) {

	method := jwt.SigningMethodHS256

	currentTime := time.Now().UTC()

	claims := jwt.RegisteredClaims{
		Issuer:    "chirpy-access",
		IssuedAt:  jwt.NewNumericDate(currentTime),
		ExpiresAt: jwt.NewNumericDate(currentTime.Add(time.Duration(expiresIn) * time.Second)),
		Subject:   userId.String(),
	}
	token := jwt.NewWithClaims(method, claims)

	signedString, err := token.SignedString([]byte(tokenSecret))
	if err != nil {
		return "", err
	}
	return signedString, nil
}

func ValidateJWT(tokenString, tokenSecret string) (uuid.UUID, error) {
	claims := &jwt.RegisteredClaims{}
	keyFunc := func(token *jwt.Token) (any, error) {
		return []byte(tokenSecret), nil
	}
	token, err := jwt.ParseWithClaims(tokenString, claims, keyFunc)

	if err != nil {
		return uuid.Nil, err
	}

	if !token.Valid {
		return uuid.Nil, errors.New("invalid JWT token")
	}

	userID, err := uuid.Parse(claims.Subject)

	if err != nil {
		return uuid.Nil, err
	}

	return userID, nil
}

func GetBearerToken(headers http.Header) (string, error) {
	authHead := headers.Get("Authorization")

	if authHead == "" || len(authHead) < 7 {
		return "", errors.New("Bearer token required")
	}

	tokenString := strings.TrimPrefix(authHead, "Bearer ")

	if len(tokenString) == len(authHead) {
		return "", errors.New("Bearer token required")
	}

	return tokenString, nil

}

func MakeRefreshToken() string {
	key := make([]byte, 32)

	rand.Read(key)

	return hex.EncodeToString(key)
}
