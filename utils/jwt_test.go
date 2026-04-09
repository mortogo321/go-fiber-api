package utils

import (
	"testing"
	"time"
)

const testSecret = "test-secret-key"

func TestGenerateToken_CreatesValidToken(t *testing.T) {
	tokenStr, err := GenerateToken(1, "user", testSecret, 1*time.Hour)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if tokenStr == "" {
		t.Fatal("expected non-empty token string")
	}
}

func TestValidateToken_ParsesClaimsCorrectly(t *testing.T) {
	tokenStr, err := GenerateToken(42, "admin", testSecret, 1*time.Hour)
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	claims, err := ValidateToken(tokenStr, testSecret)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if claims.UserID != 42 {
		t.Errorf("expected UserID 42, got %d", claims.UserID)
	}
	if claims.Role != "admin" {
		t.Errorf("expected Role 'admin', got %q", claims.Role)
	}
}

func TestValidateToken_RejectsExpiredToken(t *testing.T) {
	// Generate a token that expired 1 hour ago.
	tokenStr, err := GenerateToken(1, "user", testSecret, -1*time.Hour)
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	_, err = ValidateToken(tokenStr, testSecret)
	if err == nil {
		t.Fatal("expected error for expired token, got nil")
	}
}

func TestValidateToken_RejectsWrongSecret(t *testing.T) {
	tokenStr, err := GenerateToken(1, "user", testSecret, 1*time.Hour)
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	_, err = ValidateToken(tokenStr, "wrong-secret")
	if err == nil {
		t.Fatal("expected error for wrong secret, got nil")
	}
}

func TestValidateToken_RejectsGarbageToken(t *testing.T) {
	_, err := ValidateToken("not.a.token", testSecret)
	if err == nil {
		t.Fatal("expected error for garbage token, got nil")
	}
}
