package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"
	"regexp"

	"github.com/gin-gonic/gin"
	"github.com/kuzey/secure-deploy-platform/backend/api-gateway/internal/deployments"
)

type Handler struct {
	service *deployments.Service
}

var uuidPattern = regexp.MustCompile(`^[a-fA-F0-9-]{36}$`)

// New builds the Gin router and wires deployment endpoints to their handlers.
func New(service *deployments.Service) *gin.Engine {
	handler := &Handler{service: service}

	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery())
	router.HandleMethodNotAllowed = true
	router.NoMethod(methodNotAllowed)

	handler.registerHealthRoutes(router)
	handler.registerDeploymentRoutes(router)

	return router
}

// registerHealthRoutes registers liveness-related endpoints.
func (h *Handler) registerHealthRoutes(router *gin.Engine) {
	router.GET("/healthz", h.handleHealth)
}

// registerDeploymentRoutes registers deployment collection and detail endpoints.
func (h *Handler) registerDeploymentRoutes(router *gin.Engine) {
	deploymentAPI := router.Group("/api/deployments")
	{
		deploymentAPI.POST("", h.createDeployment)
		deploymentAPI.GET("", h.listDeployments)
		deploymentAPI.GET("/:id", h.getDeploymentByID)
	}
}

// handleHealth serves a lightweight liveness endpoint for the API process.
func (h *Handler) handleHealth(c *gin.Context) {
	writeJSON(c, http.StatusOK, gin.H{
		"status": "ok",
	})
}

// getDeploymentByID returns a single deployment resource for a valid UUID path segment.
func (h *Handler) getDeploymentByID(c *gin.Context) {
	id := c.Param("id")
	if !uuidPattern.MatchString(id) {
		writeError(c, http.StatusBadRequest, "invalid deployment id", nil)
		return
	}

	deployment, err := h.service.GetByID(c.Request.Context(), id)
	if errors.Is(err, deployments.ErrNotFound) {
		writeError(c, http.StatusNotFound, "deployment not found", nil)
		return
	}
	if err != nil {
		writeError(c, http.StatusInternalServerError, "failed to load deployment", nil)
		return
	}

	writeJSON(c, http.StatusOK, deployment)
}

// createDeployment decodes the request body, validates it, and stores a new deployment.
func (h *Handler) createDeployment(c *gin.Context) {
	defer c.Request.Body.Close()

	var req deployments.CreateRequest
	decoder := json.NewDecoder(c.Request.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&req); err != nil {
		writeError(c, http.StatusBadRequest, "invalid request body", nil)
		return
	}

	if issues := req.Normalize().Validate(); issues != nil {
		writeError(c, http.StatusBadRequest, "validation failed", issues)
		return
	}

	deployment, err := h.service.Create(c.Request.Context(), req)
	if err != nil {
		writeError(c, http.StatusInternalServerError, "failed to create deployment", nil)
		return
	}

	writeJSON(c, http.StatusCreated, deployment)
}

// listDeployments returns every stored deployment in descending creation order.
func (h *Handler) listDeployments(c *gin.Context) {
	items, err := h.service.List(c.Request.Context())
	if err != nil {
		writeError(c, http.StatusInternalServerError, "failed to list deployments", nil)
		return
	}

	writeJSON(c, http.StatusOK, gin.H{
		"items": items,
	})
}

// methodNotAllowed returns a JSON 405 response for known routes with unsupported HTTP methods.
func methodNotAllowed(c *gin.Context) {
	writeError(c, http.StatusMethodNotAllowed, "method not allowed", nil)
}

// writeJSON serializes a response payload as formatted JSON with the given status code.
func writeJSON(c *gin.Context, status int, payload any) {
	c.IndentedJSON(status, payload)
}

// writeError sends a consistent JSON error response and optional details payload.
func writeError(c *gin.Context, status int, message string, details any) {
	payload := gin.H{
		"error": message,
	}
	if details != nil {
		payload["details"] = details
	}

	writeJSON(c, status, payload)
}
