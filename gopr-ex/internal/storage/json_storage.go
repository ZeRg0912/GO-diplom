package storage

import (
	"advanced-blog-management-system/internal/errors/apperrors"
	"advanced-blog-management-system/internal/model"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

type JSONStorage struct {
	mu sync.RWMutex

	dataDir string

	users    []model.User
	posts    []model.Post
	comments []model.Comment

	nextUserID    int
	nextPostID    int
	nextCommentID int
}

func NewJSONStorage(dataDir string) (*JSONStorage, error) {
	if strings.TrimSpace(dataDir) == "" {
		dataDir = "data"
	}
	if err := os.MkdirAll(dataDir, 0o755); err != nil {
		return nil, fmt.Errorf("create data dir: %w", err)
	}
	s := &JSONStorage{
		dataDir:  dataDir,
		users:    make([]model.User, 0),
		posts:    make([]model.Post, 0),
		comments: make([]model.Comment, 0),
	}
	if err := s.load(); err != nil {
		return nil, err
	}
	s.normalizeSlices()
	s.recalculateIDs()
	return s, nil
}

func (s *JSONStorage) CreateUser(user model.User) (model.User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, existing := range s.users {
		if strings.EqualFold(existing.Email, user.Email) || strings.EqualFold(existing.Username, user.Username) {
			return model.User{}, apperrors.ErrUserAlreadyExists
		}
	}

	user.ID = s.nextUserID
	s.nextUserID++
	user.CreatedAt = time.Now().UTC()
	s.users = append(s.users, user)
	if err := s.saveUsersLocked(); err != nil {
		return model.User{}, err
	}
	return user, nil
}

func (s *JSONStorage) GetUserByID(id int) (model.User, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, user := range s.users {
		if user.ID == id {
			return user, true
		}
	}
	return model.User{}, false
}

func (s *JSONStorage) GetUserByEmail(email string) (model.User, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, user := range s.users {
		if strings.EqualFold(user.Email, email) {
			return user, true
		}
	}
	return model.User{}, false
}

func (s *JSONStorage) UserExists(email, username string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, user := range s.users {
		if strings.EqualFold(user.Email, email) || strings.EqualFold(user.Username, username) {
			return true
		}
	}
	return false
}

func (s *JSONStorage) CreatePost(post model.Post) (model.Post, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	post.ID = s.nextPostID
	s.nextPostID++
	post.CreatedAt = time.Now().UTC()
	s.posts = append(s.posts, post)
	if err := s.savePostsLocked(); err != nil {
		return model.Post{}, err
	}
	return post, nil
}

func (s *JSONStorage) GetPostByID(id int) (model.Post, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, post := range s.posts {
		if post.ID == id {
			return post, true
		}
	}
	return model.Post{}, false
}

func (s *JSONStorage) ListPosts() []model.Post {
	s.mu.RLock()
	defer s.mu.RUnlock()
	posts := append([]model.Post(nil), s.posts...)
	sort.Slice(posts, func(i, j int) bool {
		return posts[i].CreatedAt.After(posts[j].CreatedAt)
	})
	return posts
}

func (s *JSONStorage) CreateComment(comment model.Comment) (model.Comment, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	comment.ID = s.nextCommentID
	s.nextCommentID++
	comment.CreatedAt = time.Now().UTC()
	s.comments = append(s.comments, comment)
	if err := s.saveCommentsLocked(); err != nil {
		return model.Comment{}, err
	}
	return comment, nil
}

func (s *JSONStorage) ListCommentsByPostID(postID int) []model.Comment {
	s.mu.RLock()
	defer s.mu.RUnlock()
	comments := make([]model.Comment, 0)
	for _, comment := range s.comments {
		if comment.PostID == postID {
			comments = append(comments, comment)
		}
	}
	sort.Slice(comments, func(i, j int) bool {
		return comments[i].CreatedAt.Before(comments[j].CreatedAt)
	})
	return comments
}

func (s *JSONStorage) load() error {
	if err := loadJSON(s.path("users.json"), &s.users); err != nil {
		return fmt.Errorf("load users: %w", err)
	}
	if err := loadJSON(s.path("posts.json"), &s.posts); err != nil {
		return fmt.Errorf("load posts: %w", err)
	}
	if err := loadJSON(s.path("comments.json"), &s.comments); err != nil {
		return fmt.Errorf("load comments: %w", err)
	}
	return nil
}

func (s *JSONStorage) normalizeSlices() {
	if s.users == nil {
		s.users = make([]model.User, 0)
	}
	if s.posts == nil {
		s.posts = make([]model.Post, 0)
	}
	if s.comments == nil {
		s.comments = make([]model.Comment, 0)
	}
}

func (s *JSONStorage) recalculateIDs() {
	s.nextUserID = 1
	s.nextPostID = 1
	s.nextCommentID = 1
	for _, user := range s.users {
		if user.ID >= s.nextUserID {
			s.nextUserID = user.ID + 1
		}
	}
	for _, post := range s.posts {
		if post.ID >= s.nextPostID {
			s.nextPostID = post.ID + 1
		}
	}
	for _, comment := range s.comments {
		if comment.ID >= s.nextCommentID {
			s.nextCommentID = comment.ID + 1
		}
	}
}

func (s *JSONStorage) saveUsersLocked() error {
	return saveJSON(s.path("users.json"), s.users)
}

func (s *JSONStorage) savePostsLocked() error {
	return saveJSON(s.path("posts.json"), s.posts)
}

func (s *JSONStorage) saveCommentsLocked() error {
	return saveJSON(s.path("comments.json"), s.comments)
}

func (s *JSONStorage) path(name string) string {
	return filepath.Join(s.dataDir, name)
}

func loadJSON(path string, dst any) error {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return saveJSON(path, dst)
		}
		return err
	}
	if len(strings.TrimSpace(string(data))) == 0 {
		return saveJSON(path, dst)
	}
	return json.Unmarshal(data, dst)
}

func saveJSON(path string, value any) error {
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}
