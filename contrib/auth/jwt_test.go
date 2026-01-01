package auth

import (
	"testing"
	"time"
)

func TestJWTTokenGeneration(t *testing.T) {
	t.Run("generates access and refresh tokens", func(t *testing.T) {
		accessToken, refreshToken, err := GenerateToken(1)
		if err != nil {
			t.Fatalf("Failed to generate tokens: %v", err)
		}

		if accessToken == "" || refreshToken == "" {
			t.Error("Tokens should not be empty")
		}

		// Validate access token
		claims, err := ValidateToken(accessToken)
		if err != nil {
			t.Fatalf("Failed to validate access token: %v", err)
		}
		if claims.UserID != 1 || claims.Type != "access" {
			t.Errorf("Invalid access token claims: %+v", claims)
		}

		// Validate refresh token
		claims, err = ValidateToken(refreshToken)
		if err != nil {
			t.Fatalf("Failed to validate refresh token: %v", err)
		}
		if claims.UserID != 1 || claims.Type != "refresh" {
			t.Errorf("Invalid refresh token claims: %+v", claims)
		}
	})
}

func TestTokenExpiry(t *testing.T) {
	t.Run("expired token fails validation", func(t *testing.T) {
		// Manually create an expired token by setting duration to negative
		token, err := createToken(1, "test-jti", "access", -1*time.Hour)
		if err != nil {
			t.Fatalf("Failed to create expired token: %v", err)
		}

		_, err = ValidateToken(token)
		if err == nil || err.Error() != "token expired" {
			t.Errorf("Expected token expired error, got %v", err)
		}
	})
}
