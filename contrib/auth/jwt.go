package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/anuragcarret/djang-drf-go/orm/db"
	"github.com/anuragcarret/djang-drf-go/orm/queryset"
)

var secretKey = []byte("django-drf-go-secret-key-change-me")

type Header struct {
	Alg string `json:"alg"`
	Typ string `json:"typ"`
}

type Claims struct {
	UserID uint64 `json:"user_id"`
	Exp    int64  `json:"exp"`
	JTI    string `json:"jti"`
	Type   string `json:"type"` // "access" or "refresh"
}

func GenerateToken(userID uint64) (string, string, error) {
	// Simple version without database
	jti := fmt.Sprintf("%d-%d", userID, time.Now().UnixNano())
	accessToken, _ := createToken(userID, jti, "access", 1*time.Hour)
	refreshToken, _ := createToken(userID, jti, "refresh", 24*time.Hour)
	return accessToken, refreshToken, nil
}

// GenerateTokenPair generates access and refresh tokens and records the latter in DB
func GenerateTokenPair(database *db.DB, userID uint64) (string, string, error) {
	jti := fmt.Sprintf("%d-%d", userID, time.Now().UnixNano())

	accessToken, _ := createToken(userID, jti, "access", 24*time.Hour) // Demo: long access for ease or use 1h
	refreshToken, _ := createToken(userID, jti, "refresh", 7*24*time.Hour)

	claims, _ := ValidateToken(refreshToken)

	outstanding := &OutstandingToken{
		UserID: userID,
		JTI:    jti,
		Token:  refreshToken,
		Exp:    claims.Exp,
	}

	qs := queryset.NewQuerySet[*OutstandingToken](database)
	if err := qs.Create(outstanding); err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}

// RefreshToken validates a refresh token and returns a new access token
func RefreshToken(database *db.DB, refreshToken string) (string, error) {
	claims, err := ValidateToken(refreshToken)
	if err != nil {
		return "", err
	}

	if claims.Type != "refresh" {
		return "", errors.New("invalid token type")
	}

	// 1. Check if blacklisted
	blacklistQs := queryset.NewQuerySet[*BlacklistedToken](database)
	count, _ := blacklistQs.Filter(queryset.Q{"token": refreshToken}).Count()
	if count > 0 {
		return "", errors.New("token is blacklisted")
	}

	// 2. Refresh tokens are valid, issue new access token
	// Optionally rotate the refresh token here too
	accessToken, _ := createToken(claims.UserID, claims.JTI, "access", 24*time.Hour)
	return accessToken, nil
}

// BlacklistToken adds a refresh token to the blacklist
func BlacklistToken(database *db.DB, refreshToken string) error {
	claims, err := ValidateToken(refreshToken)
	if err != nil {
		return err
	}

	// Find in outstanding
	qs := queryset.NewQuerySet[*OutstandingToken](database)
	token, err := qs.Filter(queryset.Q{"jti": claims.JTI}).Get()
	if err != nil {
		return errors.New("token not found in outstanding list")
	}

	blacklist := &BlacklistedToken{
		TokenID: token.ID,
		Token:   refreshToken,
	}

	blQs := queryset.NewQuerySet[*BlacklistedToken](database)
	return blQs.Create(blacklist)
}

func createToken(userID uint64, jti, tokenType string, duration time.Duration) (string, error) {
	header := Header{Alg: "HS256", Typ: "JWT"}
	headerBytes, _ := json.Marshal(header)
	headerEncoded := base64.RawURLEncoding.EncodeToString(headerBytes)

	claims := Claims{
		UserID: userID,
		Exp:    time.Now().Add(duration).Unix(),
		JTI:    jti,
		Type:   tokenType,
	}
	claimsBytes, _ := json.Marshal(claims)
	claimsEncoded := base64.RawURLEncoding.EncodeToString(claimsBytes)

	unsignedToken := headerEncoded + "." + claimsEncoded
	signature := hmac.New(sha256.New, secretKey)
	signature.Write([]byte(unsignedToken))
	signatureEncoded := base64.RawURLEncoding.EncodeToString(signature.Sum(nil))

	return unsignedToken + "." + signatureEncoded, nil
}

func ValidateToken(tokenStr string) (*Claims, error) {
	parts := strings.Split(tokenStr, ".")
	if len(parts) != 3 {
		return nil, errors.New("invalid token format")
	}

	unsignedToken := parts[0] + "." + parts[1]
	signature, _ := base64.RawURLEncoding.DecodeString(parts[2])

	expectedSignature := hmac.New(sha256.New, secretKey)
	expectedSignature.Write([]byte(unsignedToken))

	if !hmac.Equal(signature, expectedSignature.Sum(nil)) {
		return nil, errors.New("invalid signature")
	}

	claimsBytes, _ := base64.RawURLEncoding.DecodeString(parts[1])
	var claims Claims
	if err := json.Unmarshal(claimsBytes, &claims); err != nil {
		return nil, err
	}

	if time.Now().Unix() > claims.Exp {
		return nil, errors.New("token expired")
	}

	return &claims, nil
}
