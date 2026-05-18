package handler

import (
	"advanced-blog-management-system/internal/errors/apperrors"
	"advanced-blog-management-system/internal/middleware"
	"advanced-blog-management-system/internal/model"
	"advanced-blog-management-system/internal/service"
	"encoding/json"
	"net/http"
	"strconv"
)

type PostHandler struct {
	postService *service.PostService
}

func NewPostHandler(postService *service.PostService) *PostHandler {
	return &PostHandler{postService: postService}
}

func (h *PostHandler) CreatePost(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r)
	if !ok {
		middleware.WriteError(w, apperrors.ErrUnauthorized)
		return
	}
	var req model.PostCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON body"})
		return
	}
	post, err := h.postService.CreatePost(r.Context(), &req, userID)
	if err != nil {
		middleware.WriteError(w, err)
		return
	}
	middleware.WriteJSON(w, http.StatusCreated, post)
}

func (h *PostHandler) GetPost(w http.ResponseWriter, r *http.Request, idText string) {
	id, err := strconv.Atoi(idText)
	if err != nil {
		middleware.WriteError(w, apperrors.ErrInvalidPostID)
		return
	}
	post, err := h.postService.GetPost(r.Context(), id)
	if err != nil {
		middleware.WriteError(w, err)
		return
	}
	middleware.WriteJSON(w, http.StatusOK, post)
}

func (h *PostHandler) GetAllPosts(w http.ResponseWriter, r *http.Request) {
	posts, err := h.postService.GetAllPosts(r.Context())
	if err != nil {
		middleware.WriteError(w, err)
		return
	}
	middleware.WriteJSON(w, http.StatusOK, posts)
}
