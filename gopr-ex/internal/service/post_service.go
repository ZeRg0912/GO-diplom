package service

import (
	"advanced-blog-management-system/internal/errors/apperrors"
	"advanced-blog-management-system/internal/logger"
	"advanced-blog-management-system/internal/model"
	"advanced-blog-management-system/internal/storage"
	"context"
	"fmt"
)

type PostService struct {
	store  *storage.JSONStorage
	logger *logger.AsyncLogger
}

func NewPostService(store *storage.JSONStorage, eventLogger *logger.AsyncLogger) *PostService {
	return &PostService{store: store, logger: eventLogger}
}

func (s *PostService) CreatePost(ctx context.Context, req *model.PostCreateRequest, authorID int) (model.Post, error) {
	if err := ctx.Err(); err != nil {
		return model.Post{}, err
	}
	if authorID <= 0 {
		return model.Post{}, apperrors.ErrUnauthorized
	}
	if err := req.Validate(); err != nil {
		return model.Post{}, fmt.Errorf("%w: %v", apperrors.ErrBadRequest, err)
	}
	if _, ok := s.store.GetUserByID(authorID); !ok {
		return model.Post{}, apperrors.ErrUnauthorized
	}
	post, err := s.store.CreatePost(model.Post{
		AuthorID: authorID,
		Title:    req.Title,
		Content:  req.Content,
	})
	if err != nil {
		return model.Post{}, err
	}
	s.logger.LogEvent(fmt.Sprintf("user %d created post %d", authorID, post.ID))
	return post, nil
}

func (s *PostService) GetPost(ctx context.Context, id int) (model.Post, error) {
	if err := ctx.Err(); err != nil {
		return model.Post{}, err
	}
	if id <= 0 {
		return model.Post{}, apperrors.ErrInvalidPostID
	}
	post, ok := s.store.GetPostByID(id)
	if !ok {
		return model.Post{}, apperrors.ErrPostNotFound
	}
	return post, nil
}

func (s *PostService) GetAllPosts(ctx context.Context) ([]model.Post, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return s.store.ListPosts(), nil
}
