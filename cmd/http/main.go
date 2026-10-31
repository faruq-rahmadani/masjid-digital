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

	// Wire database, cache, repositories, usecases, controller and middleware
	h, cleanup, err := newHandlers(context.Background(), cfg)
	if err != nil {
		log.Fatalf("error wiring dependencies: %v", err)
	}

	// Create and start server
	server := httpdelivery.NewHttpServer(cfg, h.controller, h.auth)
	serverErr := make(chan error, 1)
	go func() {
		serverErr <- server.ListenAndServe()
	}()

	// Run until a shutdown signal arrives or the server fails
	err = waitForShutdown(cfg, server, serverErr)
	cleanup()
	if err != nil {
		log.Fatalf("error running HTTP server: %v", err)
	}
}

// waitForShutdown blocks until SIGINT/SIGTERM, then stops the server gracefully.
// It returns early with the server's error when the server stops on its own.
func waitForShutdown(cfg *config.Schema, server httpdelivery.ServerItf, serverErr <-chan error) error {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-serverErr:
		return err
	case <-quit:
	}

	// Create shutdown context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), cfg.App.HTTPServer.GracefulTimeout)
	defer cancel()

	server.Stop(ctx)

	return nil
}
