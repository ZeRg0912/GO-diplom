package app

import (
	"advanced-blog-management-system/internal/handler"
	"advanced-blog-management-system/internal/logger"
	"advanced-blog-management-system/internal/middleware"
	"advanced-blog-management-system/internal/service"
	"advanced-blog-management-system/internal/storage"
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

type Config struct {
	ServerHost  string
	ServerPort  string
	TokenSecret string
	DataDir     string
	LogFile     string
}

type Server struct {
	authHandler    *handler.AuthHandler
	postHandler    *handler.PostHandler
	commentHandler *handler.CommentHandler
	authMiddleware *middleware.AuthMiddleware
}

func Run() error {
	_ = loadDotEnv(".env")
	cfg := loadConfig()

	store, err := storage.NewJSONStorage(cfg.DataDir)
	if err != nil {
		return fmt.Errorf("init storage: %w", err)
	}
	eventLogger, err := logger.NewAsyncLogger(cfg.LogFile, 100)
	if err != nil {
		return fmt.Errorf("init event logger: %w", err)
	}

	userService := service.NewUserService(store)
	postService := service.NewPostService(store, eventLogger)
	commentService := service.NewCommentService(store, eventLogger)

	appServer := &Server{
		authHandler:    handler.NewAuthHandler(userService, cfg.TokenSecret),
		postHandler:    handler.NewPostHandler(postService),
		commentHandler: handler.NewCommentHandler(commentService),
		authMiddleware: middleware.NewAuthMiddleware(cfg.TokenSecret),
	}

	handlerChain := middleware.LoggingMiddleware(
		middleware.RecoveryMiddleware(
			middleware.CORSMiddleware(appServer.routes()),
		),
	)

	srv := &http.Server{
		Addr:         cfg.ServerHost + ":" + cfg.ServerPort,
		Handler:      handlerChain,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	serverErr := make(chan error, 1)
	go func() {
		log.Printf("server started on http://%s", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErr <- err
		}
		close(serverErr)
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	select {
	case sig := <-stop:
		log.Printf("shutdown signal received: %s", sig)
	case err := <-serverErr:
		if err != nil {
			return err
		}
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("server shutdown: %w", err)
	}
	if err := eventLogger.Stop(shutdownCtx); err != nil {
		return fmt.Errorf("event logger shutdown: %w", err)
	}
	log.Println("server stopped")
	return nil
}

func (s *Server) routes() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := normalizePath(r.URL.Path)

		switch {
		case r.Method == http.MethodGet && path == "/health":
			handler.HealthCheckHandler(w, r)
		case r.Method == http.MethodPost && path == "/register":
			s.authHandler.RegisterHandler(w, r)
		case r.Method == http.MethodPost && path == "/login":
			s.authHandler.LoginHandler(w, r)
		case r.Method == http.MethodPost && path == "/posts":
			s.authMiddleware.Required(http.HandlerFunc(s.postHandler.CreatePost)).ServeHTTP(w, r)
		case r.Method == http.MethodGet && path == "/posts":
			s.postHandler.GetAllPosts(w, r)
		case r.Method == http.MethodGet && isPostPath(path):
			s.postHandler.GetPost(w, r, pathPart(path, 2))
		case r.Method == http.MethodPost && isCommentsPath(path):
			s.authMiddleware.Required(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				s.commentHandler.CreateComment(w, r, pathPart(path, 2))
			})).ServeHTTP(w, r)
		case r.Method == http.MethodGet && isCommentsPath(path):
			s.commentHandler.GetCommentsByPostID(w, r, pathPart(path, 2))
		default:
			if knownPath(path) {
				middleware.WriteJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
				return
			}
			middleware.WriteJSON(w, http.StatusNotFound, map[string]string{"error": "not found"})
		}
	})
}

func normalizePath(path string) string {
	if path == "" {
		return "/"
	}
	if len(path) > 1 && strings.HasSuffix(path, "/") {
		path = strings.TrimRight(path, "/")
	}
	return path
}

func isPostPath(path string) bool {
	parts := splitPath(path)
	return len(parts) == 2 && parts[0] == "posts" && parts[1] != ""
}

func knownPath(path string) bool {
	return path == "/health" || path == "/register" || path == "/login" || path == "/posts" || isPostPath(path) || isCommentsPath(path)
}

func isCommentsPath(path string) bool {
	parts := splitPath(path)
	return len(parts) == 3 && parts[0] == "posts" && parts[1] != "" && parts[2] == "comments"
}

func pathPart(path string, oneBasedIndex int) string {
	parts := splitPath(path)
	idx := oneBasedIndex - 1
	if idx < 0 || idx >= len(parts) {
		return ""
	}
	return parts[idx]
}

func splitPath(path string) []string {
	path = strings.Trim(path, "/")
	if path == "" {
		return nil
	}
	return strings.Split(path, "/")
}

func loadConfig() Config {
	return Config{
		ServerHost:  getEnv("SERVER_HOST", "0.0.0.0"),
		ServerPort:  getEnv("SERVER_PORT", "8080"),
		TokenSecret: getEnv("TOKEN_SECRET", getEnv("JWT_SECRET", "dev-secret-change-me")),
		DataDir:     getEnv("DATA_DIR", "data"),
		LogFile:     getEnv("LOG_FILE", "log.txt"),
	}
}

func getEnv(key, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
}

func loadDotEnv(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		value = strings.Trim(strings.TrimSpace(value), `"'`)
		if key == "" {
			continue
		}
		if _, exists := os.LookupEnv(key); !exists {
			_ = os.Setenv(key, value)
		}
	}
	return nil
}
