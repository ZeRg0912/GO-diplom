package auth

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"strings"
	"time"
)

type Claims struct {
	UserID    int
	Email     string
	Username  string
	ExpiresAt time.Time
}

type tokenPayload struct {
	UserID    int    `json:"user_id"`
	Email     string `json:"email"`
	Username  string `json:"username"`
	ExpiresAt int64  `json:"expires_at"`
	Nonce     string `json:"nonce"`
}

func GenerateToken(userID int, email, username, secret string) (string, time.Time, error) {
	if strings.TrimSpace(secret) == "" {
		return "", time.Time{}, errors.New("token secret is empty")
	}
	expiresAt := time.Now().Add(24 * time.Hour).UTC()
	nonce, err := randomNonce()
	if err != nil {
		return "", time.Time{}, err
	}
	payload := tokenPayload{
		UserID:    userID,
		Email:     email,
		Username:  username,
		ExpiresAt: expiresAt.Unix(),
		Nonce:     nonce,
	}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return "", time.Time{}, err
	}
	payloadPart := base64.RawURLEncoding.EncodeToString(payloadBytes)
	signaturePart := base64.RawURLEncoding.EncodeToString(sign(payloadPart, secret))
	return payloadPart + "." + signaturePart, expiresAt, nil
}

func ValidateToken(tokenString, secret string) (*Claims, error) {
	if strings.TrimSpace(secret) == "" {
		return nil, errors.New("token secret is empty")
	}
	parts := strings.Split(tokenString, ".")
	if len(parts) != 2 {
		return nil, errors.New("invalid token format")
	}
	payloadBytes, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return nil, errors.New("invalid token payload")
	}
	sig, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, errors.New("invalid token signature")
	}
	expected := sign(parts[0], secret)
	if !hmac.Equal(sig, expected) {
		return nil, errors.New("invalid token signature")
	}
	var payload tokenPayload
	if err := json.Unmarshal(payloadBytes, &payload); err != nil {
		return nil, errors.New("invalid token claims")
	}
	if payload.UserID <= 0 {
		return nil, errors.New("invalid token user id")
	}
	expiresAt := time.Unix(payload.ExpiresAt, 0).UTC()
	if !expiresAt.After(time.Now().UTC()) {
		return nil, errors.New("token expired")
	}
	return &Claims{
		UserID:    payload.UserID,
		Email:     payload.Email,
		Username:  payload.Username,
		ExpiresAt: expiresAt,
	}, nil
}

func sign(payloadPart, secret string) []byte {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(payloadPart))
	return mac.Sum(nil)
}

func randomNonce() (string, error) {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}
