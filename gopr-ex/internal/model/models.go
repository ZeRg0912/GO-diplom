package model

import (
	"errors"
	"strings"
	"time"
)

type User struct {
	ID           int       `json:"id"`
	Username     string    `json:"username"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"password_hash"`
	CreatedAt    time.Time `json:"created_at"`
}

type Post struct {
	ID        int       `json:"id"`
	AuthorID  int       `json:"author_id"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}

type Comment struct {
	ID        int       `json:"id"`
	PostID    int       `json:"post_id"`
	AuthorID  int       `json:"author_id"`
	Text      string    `json:"text"`
	CreatedAt time.Time `json:"created_at"`
}

type UserCreateRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type UserLoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type PostCreateRequest struct {
	Title   string `json:"title"`
	Content string `json:"content"`
}

type CommentCreateRequest struct {
	Text    string `json:"text"`
	Content string `json:"content,omitempty"`
}

type UserResponse struct {
	ID        int       `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}

type TokenResponse struct {
	Token     string       `json:"token"`
	ExpiresAt time.Time    `json:"expires_at"`
	User      UserResponse `json:"user"`
}

func (u User) ToResponse() UserResponse {
	return UserResponse{
		ID:        u.ID,
		Username:  u.Username,
		Email:     u.Email,
		CreatedAt: u.CreatedAt,
	}
}

func (r *UserCreateRequest) Normalize() {
	r.Username = strings.TrimSpace(r.Username)
	r.Email = strings.ToLower(strings.TrimSpace(r.Email))
}

func (r *UserCreateRequest) Validate() error {
	r.Normalize()
	if len(r.Username) < 3 || len(r.Username) > 50 {
		return errors.New("username must be between 3 and 50 characters")
	}
	if !isValidEmail(r.Email) {
		return errors.New("email must be valid")
	}
	if len(r.Password) < 6 {
		return errors.New("password must contain at least 6 characters")
	}
	return nil
}

func (r *UserLoginRequest) Normalize() {
	r.Email = strings.ToLower(strings.TrimSpace(r.Email))
}

func (r *UserLoginRequest) Validate() error {
	r.Normalize()
	if !isValidEmail(r.Email) {
		return errors.New("email must be valid")
	}
	if strings.TrimSpace(r.Password) == "" {
		return errors.New("password is required")
	}
	return nil
}

func (r *PostCreateRequest) Validate() error {
	r.Title = strings.TrimSpace(r.Title)
	r.Content = strings.TrimSpace(r.Content)
	if r.Title == "" {
		return errors.New("title is required")
	}
	if len(r.Title) > 200 {
		return errors.New("title must not exceed 200 characters")
	}
	if r.Content == "" {
		return errors.New("content is required")
	}
	return nil
}

func (r *CommentCreateRequest) Normalize() {
	r.Text = strings.TrimSpace(r.Text)
	if r.Text == "" {
		r.Text = strings.TrimSpace(r.Content)
	}
}

func (r *CommentCreateRequest) Validate() error {
	r.Normalize()
	if r.Text == "" {
		return errors.New("text is required")
	}
	if len(r.Text) > 1000 {
		return errors.New("text must not exceed 1000 characters")
	}
	return nil
}

func isValidEmail(email string) bool {
	at := strings.Index(email, "@")
	dot := strings.LastIndex(email, ".")
	return at > 0 && dot > at+1 && dot < len(email)-1
}
