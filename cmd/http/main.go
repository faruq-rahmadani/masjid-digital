package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/faruq-rahmadani/masjid-digital/internal/config"
	httpdelivery "github.com/faruq-rahmadani/masjid-digital/internal/delivery/http"
	"github.com/faruq-rahmadani/masjid-digital/pkg/env"
)

func main() {
	// Init ENV, empty falls back to development
	if err := env.Init(os.Getenv("ENV")); err != nil {
		log.Fatalf("error initializing env: %v", err)
	}

	// Load configuration of the current environment
	cfg, err := config.LoadAll(env.GetEnv())
	if err != nil {
		log.Fatalf("error loading configuration: %v", err)
	}

	log.Printf("configuration loaded: app=%s env=%s port=%s log_level=%s",
		cfg.App.Name,
		env.GetEnv(),
		cfg.App.HTTPServer.Port,
		cfg.App.Logger.Level,
	)

	// Create and start server
	server := httpdelivery.NewHttpServer(cfg)
	go server.ListenAndServe()

	// Handle graceful shutdown
	handleGracefulShutdown(cfg, server)
}

func handleGracefulShutdown(cfg *config.Schema, server httpdelivery.ServerItf) {
	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	// Create shutdown context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), cfg.App.HTTPServer.GracefulTimeout)
	defer cancel()

	server.Stop(ctx)
}
