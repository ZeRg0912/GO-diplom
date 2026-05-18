package middleware

import (
	"advanced-blog-management-system/internal/errors/apperrors"
	"advanced-blog-management-system/pkg/auth"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
)

type contextKey string

const userIDKey contextKey = "user_id"

type AuthMiddleware struct {
	secret string
}

func NewAuthMiddleware(secret string) *AuthMiddleware {
	return &AuthMiddleware{secret: secret}
}

func (m *AuthMiddleware) Required(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token, err := ExtractToken(r)
		if err != nil {
			WriteError(w, apperrors.ErrUnauthorized)
			return
		}
		claims, err := auth.ValidateToken(token, m.secret)
		if err != nil {
			WriteError(w, apperrors.ErrUnauthorized)
			return
		}
		ctx := context.WithValue(r.Context(), userIDKey, claims.UserID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func GetUserIDFromContext(r *http.Request) (int, bool) {
	userID, ok := r.Context().Value(userIDKey).(int)
	return userID, ok
}

func ExtractToken(r *http.Request) (string, error) {
	header := strings.TrimSpace(r.Header.Get("Authorization"))
	if header == "" {
		return "", errors.New("missing authorization header")
	}
	const prefix = "Bearer "
	if !strings.HasPrefix(header, prefix) {
		return "", errors.New("invalid authorization header")
	}
	token := strings.TrimSpace(strings.TrimPrefix(header, prefix))
	if token == "" {
		return "", errors.New("empty bearer token")
	}
	return token, nil
}

func WriteJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func WriteError(w http.ResponseWriter, err error) {
	WriteJSON(w, apperrors.StatusCode(err), map[string]string{"error": err.Error()})
}
