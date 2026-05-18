package apperrors

import (
	"errors"
	"net/http"
)

var (
	ErrBadRequest         = errors.New("bad request")
	ErrUnauthorized       = errors.New("unauthorized")
	ErrForbidden          = errors.New("forbidden")
	ErrUserNotFound       = errors.New("user not found")
	ErrUserAlreadyExists  = errors.New("user already exists")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrPostNotFound       = errors.New("post not found")
	ErrCommentNotFound    = errors.New("comment not found")
	ErrInvalidPostID      = errors.New("invalid post id")
	ErrInvalidCommentID   = errors.New("invalid comment id")
)

func StatusCode(err error) int {
	switch {
	case err == nil:
		return http.StatusOK
	case errors.Is(err, ErrUnauthorized):
		return http.StatusUnauthorized
	case errors.Is(err, ErrForbidden):
		return http.StatusForbidden
	case errors.Is(err, ErrUserNotFound), errors.Is(err, ErrPostNotFound), errors.Is(err, ErrCommentNotFound):
		return http.StatusNotFound
	case errors.Is(err, ErrUserAlreadyExists):
		return http.StatusConflict
	case errors.Is(err, ErrInvalidCredentials):
		return http.StatusUnauthorized
	case errors.Is(err, ErrBadRequest), errors.Is(err, ErrInvalidPostID), errors.Is(err, ErrInvalidCommentID):
		return http.StatusBadRequest
	default:
		return http.StatusInternalServerError
	}
}
