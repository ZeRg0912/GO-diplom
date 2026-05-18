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

type CommentHandler struct {
	commentService *service.CommentService
}

func NewCommentHandler(commentService *service.CommentService) *CommentHandler {
	return &CommentHandler{commentService: commentService}
}

func (h *CommentHandler) CreateComment(w http.ResponseWriter, r *http.Request, postIDText string) {
	userID, ok := middleware.GetUserIDFromContext(r)
	if !ok {
		middleware.WriteError(w, apperrors.ErrUnauthorized)
		return
	}
	postID, err := strconv.Atoi(postIDText)
	if err != nil {
		middleware.WriteError(w, apperrors.ErrInvalidPostID)
		return
	}
	var req model.CommentCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON body"})
		return
	}
	comment, err := h.commentService.CreateComment(r.Context(), &req, postID, userID)
	if err != nil {
		middleware.WriteError(w, err)
		return
	}
	middleware.WriteJSON(w, http.StatusCreated, comment)
}

func (h *CommentHandler) GetCommentsByPostID(w http.ResponseWriter, r *http.Request, postIDText string) {
	postID, err := strconv.Atoi(postIDText)
	if err != nil {
		middleware.WriteError(w, apperrors.ErrInvalidPostID)
		return
	}
	comments, err := h.commentService.GetCommentsByPostID(r.Context(), postID)
	if err != nil {
		middleware.WriteError(w, err)
		return
	}
	middleware.WriteJSON(w, http.StatusOK, comments)
}
