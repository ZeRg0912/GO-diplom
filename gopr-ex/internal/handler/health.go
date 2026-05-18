package handler

import (
	"advanced-blog-management-system/internal/middleware"
	"net/http"
)

func HealthCheckHandler(w http.ResponseWriter, r *http.Request) {
	middleware.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
