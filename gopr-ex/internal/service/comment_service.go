package service

import (
	"advanced-blog-management-system/internal/errors/apperrors"
	"advanced-blog-management-system/internal/logger"
	"advanced-blog-management-system/internal/model"
	"advanced-blog-management-system/internal/storage"
	"context"
	"fmt"
)

type CommentService struct {
	store  *storage.JSONStorage
	logger *logger.AsyncLogger
}

func NewCommentService(store *storage.JSONStorage, eventLogger *logger.AsyncLogger) *CommentService {
	return &CommentService{store: store, logger: eventLogger}
}

func (s *CommentService) CreateComment(ctx context.Context, req *model.CommentCreateRequest, postID int, authorID int) (model.Comment, error) {
	if err := ctx.Err(); err != nil {
		return model.Comment{}, err
	}
	if authorID <= 0 {
		return model.Comment{}, apperrors.ErrUnauthorized
	}
	if postID <= 0 {
		return model.Comment{}, apperrors.ErrInvalidPostID
	}
	if err := req.Validate(); err != nil {
		return model.Comment{}, fmt.Errorf("%w: %v", apperrors.ErrBadRequest, err)
	}
	if _, ok := s.store.GetUserByID(authorID); !ok {
		return model.Comment{}, apperrors.ErrUnauthorized
	}
	if _, ok := s.store.GetPostByID(postID); !ok {
		return model.Comment{}, apperrors.ErrPostNotFound
	}
	comment, err := s.store.CreateComment(model.Comment{
		PostID:   postID,
		AuthorID: authorID,
		Text:     req.Text,
	})
	if err != nil {
		return model.Comment{}, err
	}
	s.logger.LogEvent(fmt.Sprintf("user %d created comment %d on post %d", authorID, comment.ID, postID))
	return comment, nil
}

func (s *CommentService) GetCommentsByPostID(ctx context.Context, postID int) ([]model.Comment, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if postID <= 0 {
		return nil, apperrors.ErrInvalidPostID
	}
	if _, ok := s.store.GetPostByID(postID); !ok {
		return nil, apperrors.ErrPostNotFound
	}
	return s.store.ListCommentsByPostID(postID), nil
}
