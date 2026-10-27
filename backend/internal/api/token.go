package api

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"strings"
	"time"
)

const (
	tokenIssuer   = "fullstack-app"
	tokenAudience = "fullstack-app-api"
	tokenLifetime = 12 * time.Hour
)

type tokenHeader struct {
	Algorithm string `json:"alg"`
	Type      string `json:"typ"`
}

type TokenClaims struct {
	Subject   string `json:"sub"`
	Issuer    string `json:"iss"`
	Audience  string `json:"aud"`
	IssuedAt  int64  `json:"iat"`
	ExpiresAt int64  `json:"exp"`
}

func IssueToken(secret []byte, username string) (string, error) {
	now := time.Now()
	header, err := json.Marshal(tokenHeader{Algorithm: "HS256", Type: "JWT"})
	if err != nil {
		return "", err
	}
	claims, err := json.Marshal(TokenClaims{
		Subject: username, Issuer: tokenIssuer, Audience: tokenAudience,
		IssuedAt: now.Unix(), ExpiresAt: now.Add(tokenLifetime).Unix(),
	})
	if err != nil {
		return "", err
	}
	unsigned := base64.RawURLEncoding.EncodeToString(header) + "." +
		base64.RawURLEncoding.EncodeToString(claims)
	return unsigned + "." + sign(secret, unsigned), nil
}

func validateToken(secret []byte, token string) bool {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return false
	}
	unsigned := parts[0] + "." + parts[1]
	expectedSignature := sign(secret, unsigned)
	if subtle.ConstantTimeCompare([]byte(parts[2]), []byte(expectedSignature)) != 1 {
		return false
	}
	headerJSON, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return false
	}
	var header tokenHeader
	if json.Unmarshal(headerJSON, &header) != nil || header.Algorithm != "HS256" || header.Type != "JWT" {
		return false
	}
	claimsJSON, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return false
	}
	var claims TokenClaims
	if json.Unmarshal(claimsJSON, &claims) != nil {
		return false
	}
	now := time.Now().Unix()
	return claims.Subject != "" && claims.Issuer == tokenIssuer &&
		claims.Audience == tokenAudience && claims.IssuedAt <= now &&
		claims.ExpiresAt > now && claims.ExpiresAt > claims.IssuedAt
}

func sign(secret []byte, message string) string {
	mac := hmac.New(sha256.New, secret)
	_, _ = mac.Write([]byte(message))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}

func credentialsMatch(provided, expected string) bool {
	providedHash := sha256.Sum256([]byte(provided))
	expectedHash := sha256.Sum256([]byte(expected))
	return subtle.ConstantTimeCompare(providedHash[:], expectedHash[:]) == 1 &&
		len(provided) == len(expected)
}
