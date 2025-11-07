package httpserver

import (
	"context"
	"net/http"

	"github.com/deepakvbansode/idp-cloudgenie-backend/internal/config"
	"github.com/deepakvbansode/idp-cloudgenie-backend/internal/core/ports"
	"github.com/deepakvbansode/idp-cloudgenie-backend/internal/core/usecases"
)

type Server struct {
	logger ports.Logger
	Config 		 	*config.WebServerConfig
	Router           *Router
	BlueprintService *usecases.BlueprintService
	ResourceService  *usecases.ResourceService
}

func NewServer(logger ports.Logger, config *config.WebServerConfig, blueprintService *usecases.BlueprintService, resourceService *usecases.ResourceService) *Server {
	r := NewRouter(logger, config, blueprintService, resourceService)
	r.InitializeRouter()
	return &Server{
		logger:          logger,
		Config: 		 config,
		Router:           r,
		
	}
}

// Start runs the HTTP server and supports context-based shutdown
func (s *Server) Start(ctx context.Context) error {
       srv := &http.Server{
	       Addr:    ":" + s.Config.Port,
	       Handler: s.Router,
       }
       go func() {
	       <-ctx.Done()
	       shutdownCtx, cancel := context.WithTimeout(context.Background(), 5_000_000_000) // 5s
	       defer cancel()
	       if err := srv.Shutdown(shutdownCtx); err != nil {
		       s.logger.Error("HTTP server shutdown error: ", err)
	       }
       }()
       s.logger.Info("Running server on port: ", s.Config.Port)
       err := srv.ListenAndServe()
       if err != nil && err != http.ErrServerClosed {
	       s.logger.Error("HTTP server error: ", err)
	       return err
       }
       return nil
}
