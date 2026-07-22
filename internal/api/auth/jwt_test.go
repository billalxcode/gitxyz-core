package auth

import (
	"testing"
	"time"

	"gitxyz/internal/models"

	"github.com/google/uuid"
)

func TestGenerateAndParseToken(t *testing.T) {
	user := &models.User{Base: models.Base{ID: uuid.New()}, Username: "alice", Email: "alice@example.com"}

	token, err := GenerateToken(user, "access")
	if err != nil {
		t.Fatalf("expected token generation to succeed, got error: %v", err)
	}

	claims, err := ParseToken(token)
	if err != nil {
		t.Fatalf("expected token parsing to succeed, got error: %v", err)
	}

	if claims.UserID != user.ID.String() {
		t.Fatalf("expected user id %s, got %s", user.ID.String(), claims.UserID)
	}

	if claims.TokenType != "access" {
		t.Fatalf("expected token type access, got %s", claims.TokenType)
	}

	if claims.ExpiresAt == nil || claims.ExpiresAt.Time.Before(time.Now()) {
		t.Fatal("expected token to have a future expiry")
	}
}

func TestRevokeTokenBlocksToken(t *testing.T) {
	user := &models.User{Base: models.Base{ID: uuid.New()}, Username: "bob", Email: "bob@example.com"}

	token, err := GenerateToken(user, "refresh")
	if err != nil {
		t.Fatalf("expected refresh token generation to succeed, got error: %v", err)
	}

	RevokeToken(token)

	if !IsTokenRevoked(token) {
		t.Fatal("expected token to be marked as revoked")
	}
}
