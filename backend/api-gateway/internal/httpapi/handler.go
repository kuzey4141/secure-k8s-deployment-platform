package httpapi

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/kuzey/secure-deploy-platform/backend/api-gateway/internal/deployments"
)

type Handler struct {
	service *deployments.Service
}

var uuidPattern = regexp.MustCompile(`^[a-fA-F0-9-]{36}$`)

func New(service *deployments.Service) http.Handler {
	handler := &Handler{service: service}

	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", handler.handleHealth)
	mux.HandleFunc("/api/deployments", handler.handleDeployments)
	mux.HandleFunc("/api/deployments/", handler.handleDeploymentByID)

	return loggingMiddleware(mux)
}

func (h *Handler) handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed", nil)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"status": "ok",
	})
}

func (h *Handler) handleDeployments(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		h.createDeployment(w, r)
	case http.MethodGet:
		h.listDeployments(w, r)
	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed", nil)
	}
}

func (h *Handler) handleDeploymentByID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed", nil)
		return
	}

	id := strings.TrimPrefix(r.URL.Path, "/api/deployments/")
	if id == "" || strings.Contains(id, "/") {
		writeError(w, http.StatusNotFound, "deployment not found", nil)
		return
	}
	if !uuidPattern.MatchString(id) {
		writeError(w, http.StatusBadRequest, "invalid deployment id", nil)
		return
	}

	deployment, err := h.service.GetByID(r.Context(), id)
	if errors.Is(err, deployments.ErrNotFound) {
		writeError(w, http.StatusNotFound, "deployment not found", nil)
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to load deployment", nil)
		return
	}

	writeJSON(w, http.StatusOK, deployment)
}

func (h *Handler) createDeployment(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	var req deployments.CreateRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body", nil)
		return
	}

	if issues := req.Normalize().Validate(); issues != nil {
		writeError(w, http.StatusBadRequest, "validation failed", issues)
		return
	}

	deployment, err := h.service.Create(r.Context(), req)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create deployment", nil)
		return
	}

	writeJSON(w, http.StatusCreated, deployment)
}

func (h *Handler) listDeployments(w http.ResponseWriter, r *http.Request) {
	items, err := h.service.List(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list deployments", nil)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"items": items,
	})
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	_ = encoder.Encode(payload)
}

func writeError(w http.ResponseWriter, status int, message string, details any) {
	payload := map[string]any{
		"error": message,
	}
	if details != nil {
		payload["details"] = details
	}

	writeJSON(w, status, payload)
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s %s", r.Method, r.URL.Path, time.Since(start))
	})
}
