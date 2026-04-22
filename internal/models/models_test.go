package models

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestUserJSON(t *testing.T) {
	user := User{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Email:     "test@example.com",
	}

	data, err := json.Marshal(user)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var unmarshaled User
	err = json.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if unmarshaled.ID != user.ID {
		t.Errorf("ID mismatch")
	}
	if unmarshaled.Email != user.Email {
		t.Errorf("Email mismatch")
	}
}

func TestUserParamsJSON(t *testing.T) {
	params := UserParams{
		Email:    "test@example.com",
		Password: "password",
	}

	data, err := json.Marshal(params)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var unmarshaled UserParams
	err = json.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if unmarshaled.Email != params.Email {
		t.Errorf("Email mismatch")
	}
	if unmarshaled.Password != params.Password {
		t.Errorf("Password mismatch")
	}
}

func TestChirpJSON(t *testing.T) {
	chirp := Chirp{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Body:      "test body",
		UserId:    uuid.New(),
	}

	data, err := json.Marshal(chirp)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var unmarshaled Chirp
	err = json.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if unmarshaled.ID != chirp.ID {
		t.Errorf("ID mismatch")
	}
	if unmarshaled.Body != chirp.Body {
		t.Errorf("Body mismatch")
	}
	if unmarshaled.UserId != chirp.UserId {
		t.Errorf("UserId mismatch")
	}
}
