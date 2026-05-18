package handler

import (
	"advanced-blog-management-system/internal/middleware"
	"advanced-blog-management-system/internal/model"
	"advanced-blog-management-system/internal/service"
	"advanced-blog-management-system/pkg/auth"
	"encoding/json"
	"fmt"
	"net/http"
)

type AuthHandler struct {
	userService *service.UserService
	jwtSecret   string
}

func NewAuthHandler(userService *service.UserService, jwtSecret string) *AuthHandler {
	return &AuthHandler{userService: userService, jwtSecret: jwtSecret}
}

func (h *AuthHandler) RegisterHandler(w http.ResponseWriter, r *http.Request) {
	var req model.UserCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON body"})
		return
	}
	user, err := h.userService.Register(r.Context(), &req)
	if err != nil {
		middleware.WriteError(w, err)
		return
	}
	resp, err := h.tokenResponse(user)
	if err != nil {
		middleware.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to generate token"})
		return
	}
	middleware.WriteJSON(w, http.StatusCreated, resp)
}

func (h *AuthHandler) LoginHandler(w http.ResponseWriter, r *http.Request) {
	var req model.UserLoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON body"})
		return
	}
	user, err := h.userService.Login(r.Context(), &req)
	if err != nil {
		middleware.WriteError(w, err)
		return
	}
	resp, err := h.tokenResponse(user)
	if err != nil {
		middleware.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to generate token"})
		return
	}
	middleware.WriteJSON(w, http.StatusOK, resp)
}

func (h *AuthHandler) tokenResponse(user model.User) (model.TokenResponse, error) {
	token, expiresAt, err := auth.GenerateToken(user.ID, user.Email, user.Username, h.jwtSecret)
	if err != nil {
		return model.TokenResponse{}, fmt.Errorf("generate token: %w", err)
	}
	return model.TokenResponse{Token: token, ExpiresAt: expiresAt, User: user.ToResponse()}, nil
}
