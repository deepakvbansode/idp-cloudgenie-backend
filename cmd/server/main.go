package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/deepakvbansode/idp-cloudgenie-backend/internal/adaptors/driven/crossplane"
	"github.com/deepakvbansode/idp-cloudgenie-backend/internal/adaptors/driven/github"
	mongodb "github.com/deepakvbansode/idp-cloudgenie-backend/internal/adaptors/driven/mongo"
	httpserver "github.com/deepakvbansode/idp-cloudgenie-backend/internal/adaptors/driver/http-server"
	k8swatcher "github.com/deepakvbansode/idp-cloudgenie-backend/internal/adaptors/driver/k8s-watcher"
	"github.com/deepakvbansode/idp-cloudgenie-backend/internal/common/logger"
	"github.com/deepakvbansode/idp-cloudgenie-backend/internal/config"
	"github.com/deepakvbansode/idp-cloudgenie-backend/internal/core/usecases"
)

	func main() {
		config, err := config.GetEnvConfig()
		if err != nil {
			panic("No config")
		}
		logger := logger.NewLogger(config.LogLevel)
		logger.Info("Starting CloudGenie Backend Service", nil)
		crossplaneAdaptor := crossplane.NewCrossplaneAdaptor(logger, config.Crossplane)
		if err != nil {
			logger.Panic("Failed to initialize Crossplane adaptor: %v", err)
		}
		blueprintService := usecases.NewBlueprintService(logger, crossplaneAdaptor)
		githubAdaptor := github.NewGithubAdaptor(logger, config.Github)
		if err != nil {
			logger.Panic("Failed to initialize Github adaptor: %v", err)
		}
		mongoRepository := mongodb.NewRepositoryAdaptor(logger, config.Mongo)
		if err != nil {
			logger.Panic("Failed to initialize MongoDB repository: %v", err)
		}
		resourceService := usecases.NewResourceService(logger, githubAdaptor, mongoRepository, crossplaneAdaptor)
		server := httpserver.NewServer(logger, config, blueprintService, resourceService)

		ctx, cancel := context.WithCancel(context.Background())
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

		// Start watcher
		watcherDone := make(chan struct{})
		go func() {
			err := k8swatcher.WatchXRDInstances(ctx, logger)
			if err != nil {
				logger.Error("K8s watcher error: ", err)
			}
			close(watcherDone)
		}()

		// Start HTTP server
		serverDone := make(chan struct{})
		go func() {
			if err := server.Start(ctx); err != nil {
				logger.Error("HTTP server error: ", err)
			}
			close(serverDone)
		}()

		// Wait for signal
		<-sigCh
		logger.Info("Shutdown signal received, shutting down...")
		cancel()

		// Wait for both goroutines to finish
		<-watcherDone
		<-serverDone
		logger.Info("Shutdown complete.")
	}