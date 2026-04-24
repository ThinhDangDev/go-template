package auth

import (
	"testing"
	"time"
)

func TestTokenManagerIssueAndValidate(t *testing.T) {
	manager := NewTokenManager("secret", "issuer", time.Minute)

	token, err := manager.IssueAccessToken("user-1", "admin@example.com", "admin")
	if err != nil {
		t.Fatalf("IssueAccessToken() error = %v", err)
	}

	claims, err := manager.ValidateToken(token)
	if err != nil {
		t.Fatalf("ValidateToken() error = %v", err)
	}

	if claims.UserID != "user-1" {
		t.Fatalf("expected user id user-1, got %s", claims.UserID)
	}

	if claims.Email != "admin@example.com" {
		t.Fatalf("expected email admin@example.com, got %s", claims.Email)
	}

	if claims.Role != "admin" {
		t.Fatalf("expected role admin, got %s", claims.Role)
	}
}

func TestTokenManagerRejectsWrongIssuer(t *testing.T) {
	issuerA := NewTokenManager("secret", "issuer-a", time.Minute)
	issuerB := NewTokenManager("secret", "issuer-b", time.Minute)

	token, err := issuerA.IssueAccessToken("user-1", "admin@example.com", "admin")
	if err != nil {
		t.Fatalf("IssueAccessToken() error = %v", err)
	}

	if _, err := issuerB.ValidateToken(token); err == nil {
		t.Fatalf("expected issuer validation to fail")
	}
}
