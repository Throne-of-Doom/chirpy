package auth

import (
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestMakeAndValidateJWT(t *testing.T) {
	// arrange
	userID := uuid.New()
	tokenSecret := "super-secret"
	expiresIn := time.Hour

	// act
	token, err := MakeJWT(userID, tokenSecret, expiresIn)
	if err != nil {
		t.Fatalf("MakeJWT returned error: %v", err)
	}

	gotID, err := ValidateJWT(token, tokenSecret)
	if err != nil {
		t.Fatalf("ValidateJWT returned error: %v", err)
	}

	// assert
	if gotID != userID {
		t.Errorf("ValidateJWT returned wrong userID: got %v, want %v", gotID, userID)
	}
}

func TestExpiredJWT(t *testing.T) {
	userID := uuid.New()
	tokenSecret := "super-secret"
	expiresIn := -time.Hour // already expired

	token, err := MakeJWT(userID, tokenSecret, expiresIn)
	if err != nil {
		t.Fatalf("MakeJWT returned error: %v", err)
	}

	gotID, err := ValidateJWT(token, tokenSecret)
	if err == nil {
		t.Fatalf("expected error for expired token, got nil (userID=%v)", gotID)
	}
}
func TestWrongSecretJWT(t *testing.T) {
	userID := uuid.New()
	goodSecret := "super-secret"
	badSecret := "stuper-secret"
	expiresIn := time.Hour

	token, err := MakeJWT(userID, goodSecret, expiresIn)
	if err != nil {
		t.Fatalf("MakeJWT returned error: %v", err)
	}

	gotID, err := ValidateJWT(token, badSecret)
	if err == nil {
		t.Fatalf("expected error when validating with wrong secret, got nil (userID=%v)", gotID)
	}
}

func TestGetBearerToken_Valid(t *testing.T) {
	headers := http.Header{}
	headers.Set("Authorization", "Bearer abc123")

	token, err := GetBearerToken(headers)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if token != "abc123" {
		t.Fatalf("expected token %q, got %q", "abc123", token)
	}
}

func TestMissingHeader(t *testing.T) {
	headers := http.Header{}
	token, err := GetBearerToken(headers)
	if err == nil {
		t.Fatalf("expected an error, got nil")
	}
	if token != "" {
		t.Fatalf("expected empty token, got %q", token)
	}
}

func TestValidHeader(t *testing.T) {
	headers := http.Header{}
	headers.Set("Authorization", "Bearer abc123")
	token, err := GetBearerToken(headers)
	if err != nil {
		t.Fatalf("expected no errors, got %v", err)
	}
	if token != "abc123" {
		t.Fatalf("expected token %q, got %q", "abc123", token)
	}
}
