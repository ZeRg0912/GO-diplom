package service

import (
	"advanced-blog-management-system/internal/errors/apperrors"
	"advanced-blog-management-system/internal/model"
	"advanced-blog-management-system/internal/storage"
	"advanced-blog-management-system/pkg/auth"
	"context"
	"fmt"
)

type UserService struct {
	store *storage.JSONStorage
}

func NewUserService(store *storage.JSONStorage) *UserService {
	return &UserService{store: store}
}

func (s *UserService) Register(ctx context.Context, req *model.UserCreateRequest) (model.User, error) {
	if err := ctx.Err(); err != nil {
		return model.User{}, err
	}
	if err := req.Validate(); err != nil {
		return model.User{}, fmt.Errorf("%w: %v", apperrors.ErrBadRequest, err)
	}
	if s.store.UserExists(req.Email, req.Username) {
		return model.User{}, apperrors.ErrUserAlreadyExists
	}
	hash, err := auth.HashPassword(req.Password)
	if err != nil {
		return model.User{}, fmt.Errorf("hash password: %w", err)
	}
	user := model.User{
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: hash,
	}
	created, err := s.store.CreateUser(user)
	if err != nil {
		return model.User{}, err
	}
	return created, nil
}

func (s *UserService) Login(ctx context.Context, req *model.UserLoginRequest) (model.User, error) {
	if err := ctx.Err(); err != nil {
		return model.User{}, err
	}
	if err := req.Validate(); err != nil {
		return model.User{}, fmt.Errorf("%w: %v", apperrors.ErrBadRequest, err)
	}
	user, ok := s.store.GetUserByEmail(req.Email)
	if !ok {
		return model.User{}, apperrors.ErrInvalidCredentials
	}
	if !auth.VerifyPassword(user.PasswordHash, req.Password) {
		return model.User{}, apperrors.ErrInvalidCredentials
	}
	return user, nil
}

func (s *UserService) GetByID(ctx context.Context, id int) (model.User, error) {
	if err := ctx.Err(); err != nil {
		return model.User{}, err
	}
	user, ok := s.store.GetUserByID(id)
	if !ok {
		return model.User{}, apperrors.ErrUserNotFound
	}
	return user, nil
}
