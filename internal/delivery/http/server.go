package http

import (
	"context"
	"errors"
	"log"
	"net/http"

	"github.com/faruq-rahmadani/masjid-digital/internal/config"
	"github.com/faruq-rahmadani/masjid-digital/internal/delivery/http/controller"
	"github.com/faruq-rahmadani/masjid-digital/internal/delivery/http/middleware"
	"github.com/labstack/echo/v4"
	echomiddleware "github.com/labstack/echo/v4/middleware"
)

type ServerItf interface {
	ListenAndServe() error
	Stop(ctx context.Context)
}

type Server struct {
	echo *echo.Echo
}

// NewHttpServer creates a new HTTP server instance serving ctrl's handlers,
// with auth guarding the routes that need a logged in user
func NewHttpServer(cfg *config.Schema, ctrl *controller.Controller, auth *middleware.Auth) ServerItf {
	e := newEcho(ctrl, auth)
	setServerObj(e, cfg.App.HTTPServer)

	return &Server{echo: e}
}

// newEcho returns the Echo instance with its middlewares and routes
func newEcho(ctrl *controller.Controller, auth *middleware.Auth) *echo.Echo {
	e := echo.New()
	e.HideBanner = true
	e.HidePort = true
	e.HTTPErrorHandler = newErrorHandler()

	e.Use(echomiddleware.Recover())
	e.Use(echomiddleware.RequestID())
	e.Use(echomiddleware.BodyLimitWithConfig(echomiddleware.BodyLimitConfig{
		Limit: defaultBodyLimit,
		// uploads set their own, larger limit on the route
		Skipper: func(c echo.Context) bool {
			return largeBodyRoutes[c.Path()]
		},
	}))

	registerRouters(e, ctrl, auth)

	return e
}

// ListenAndServe starts the HTTP server and blocks until it is stopped.
// It returns nil after Stop, or the error that kept the server from running,
// e.g. the port is already in use.
func (s *Server) ListenAndServe() error {
	log.Printf("starting HTTP server on %s", s.echo.Server.Addr)

	if err := s.echo.StartServer(s.echo.Server); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	return nil
}

// Stop gracefully shuts down the server
func (s *Server) Stop(ctx context.Context) {
	log.Print("shutting down HTTP server")

	if err := s.echo.Shutdown(ctx); err != nil {
		log.Printf("error shutting down HTTP server: %v", err)
		return
	}

	log.Print("HTTP server shutdown complete")
}

// setServerObj configures server-specific parameters
func setServerObj(e *echo.Echo, serverCfg config.HTTPServer) {
	e.Server.Addr = serverCfg.ListenAddress + ":" + serverCfg.Port
	if serverCfg.ReadTimeout > 0 {
		e.Server.ReadTimeout = serverCfg.ReadTimeout
	}
	if serverCfg.WriteTimeout > 0 {
		e.Server.WriteTimeout = serverCfg.WriteTimeout
	}
	if serverCfg.IdleTimeout > 0 {
		e.Server.IdleTimeout = serverCfg.IdleTimeout
	}
}
