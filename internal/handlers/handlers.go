package handlers

import (
	"log"
	"net/http"

	"github.com/deepakvbansode/idp-cloudgenie-backend/internal/models"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	orchestration *OrchestrationService
}

func NewHandler(orchestration *OrchestrationService) *Handler {
	return &Handler{
		orchestration: orchestration,
	}
}

// ChatHandler handles chat requests from the frontend
func (h *Handler) ChatHandler(c *gin.Context) {
	var request models.ChatRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_request",
			Message: err.Error(),
			Code:    http.StatusBadRequest,
		})
		return
	}

	log.Printf("Received chat request: %s (provider: %s, model: %s)", 
		request.Prompt, request.Provider, request.Model)

	// Process the prompt through orchestration
	response, err := h.orchestration.ProcessPrompt(c.Request.Context(), &request)
	if err != nil {
		log.Printf("Error processing prompt: %v", err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "processing_error",
			Message: err.Error(),
			Code:    http.StatusInternalServerError,
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// HealthHandler checks the health of the service
func (h *Handler) HealthHandler(c *gin.Context) {
	services := h.orchestration.HealthCheck(c.Request.Context())
	
	mcpReady := services["mcp_client"] == "connected"
	status := "healthy"
	if !mcpReady {
		status = "degraded"
	}

	c.JSON(http.StatusOK, models.HealthResponse{
		Status:         status,
		MCPServerReady: mcpReady,
		Services:       services,
	})
}

// ToolsHandler returns the list of available tools
func (h *Handler) ToolsHandler(c *gin.Context) {
	tools := h.orchestration.GetAvailableTools()
	c.JSON(http.StatusOK, models.ToolsResponse{
		Tools: tools,
	})
}

// SetupRoutes configures all HTTP routes
func SetupRoutes(router *gin.Engine, handler *Handler) {
	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		// Chat endpoint
		v1.POST("/chat", handler.ChatHandler)
		
		// Health check
		v1.GET("/health", handler.HealthHandler)
		
		// List available tools
		v1.GET("/tools", handler.ToolsHandler)
	}

	// Root health check
	router.GET("/health", handler.HealthHandler)
}
