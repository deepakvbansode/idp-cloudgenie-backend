package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/deepakvbansode/idp-cloudgenie-backend/internal/ai"
	"github.com/deepakvbansode/idp-cloudgenie-backend/internal/config"
	"github.com/deepakvbansode/idp-cloudgenie-backend/internal/handlers"
	"github.com/deepakvbansode/idp-cloudgenie-backend/internal/mcp"
	"github.com/gin-gonic/gin"
	"github.com/rs/cors"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	log.Println("Starting CloudGenie Backend Service...")
	log.Printf("AI Provider: %s", cfg.DefaultAIProvider)
	log.Printf("MCP Server URL: %s", cfg.MCPServerURL)
	log.Printf("CloudGenie Backend URL: %s", cfg.CloudGenieBackendURL)

	// Initialize MCP Client
	log.Println("Initializing MCP client...")
	mcpEnv := []string{
		fmt.Sprintf("CLOUDGENIE_BACKEND_URL=%s", cfg.CloudGenieBackendURL),
	}
	
	mcpClient, err := mcp.NewClient(cfg.MCPServerURL, mcpEnv)
	if err != nil {
		log.Fatalf("Failed to create MCP client: %v", err)
	}
	defer mcpClient.Close()

	// Initialize AI Provider
	log.Printf("Initializing AI provider: %s", cfg.DefaultAIProvider)
	var aiProvider ai.Provider
	
	if cfg.DefaultAIProvider == "openai" || cfg.DefaultAIProvider == "" {
		aiProvider, err = ai.NewOpenAIProvider(cfg.OpenAIAPIKey, cfg.OpenAIModel)
	} else if cfg.DefaultAIProvider == "anthropic" {
		aiProvider, err = ai.NewAnthropicProvider(cfg.AnthropicAPIKey, cfg.AnthropicModel)
	} else if cfg.DefaultAIProvider == "gemini" {
		aiProvider, err = ai.NewGeminiProvider(cfg.GeminiAPIKey, cfg.GeminiModel)
	} else {
		log.Fatalf("Unsupported AI provider: %s", cfg.DefaultAIProvider)
	}

	if err != nil {
		log.Fatalf("Failed to initialize AI provider: %v", err)
	}

	// Initialize Orchestration Service
	log.Println("Initializing orchestration service...")
	orchestration, err := handlers.NewOrchestrationService(mcpClient, aiProvider)
	if err != nil {
		log.Fatalf("Failed to initialize orchestration service: %v", err)
	}

	// Initialize HTTP handler
	handler := handlers.NewHandler(orchestration)

	// Setup Gin router
	if cfg.ServerHost != "localhost" && cfg.ServerHost != "127.0.0.1" {
		gin.SetMode(gin.ReleaseMode)
	}
	
	router := gin.Default()

	// Setup CORS
	corsHandler := cors.New(cors.Options{
		AllowedOrigins:   cfg.AllowedOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	})

	router.Use(func(c *gin.Context) {
		corsHandler.HandlerFunc(c.Writer, c.Request)
		c.Next()
	})

	// Setup routes
	handlers.SetupRoutes(router, handler)

	// Start server
	address := fmt.Sprintf("%s:%s", cfg.ServerHost, cfg.ServerPort)
	log.Printf("Server starting on %s", address)
	log.Println("Available endpoints:")
	log.Println("  POST /api/v1/chat       - Send chat prompts")
	log.Println("  GET  /api/v1/health     - Health check")
	log.Println("  GET  /api/v1/tools      - List available tools")

	// Setup graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := router.Run(address); err != nil {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal
	<-quit
	log.Println("Shutting down server...")
	
	// Cleanup
	mcpClient.Close()
	log.Println("Server stopped")
}
