package http

import (
	"context"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/syauqiahmd/masjid-digital-site/server/internal/config"
	"github.com/syauqiahmd/masjid-digital-site/server/internal/delivery/http/controller"
)

type Server struct {
	server *http.Server
}

func NewServer(cfg config.HTTPServer, ctrl *controller.Controller) *Server {
	return &Server{server: &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           NewRouter(ctrl),
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       120 * time.Second,
	}}
}

// ListenAndServe blocks until the server stops; a graceful Stop is not an error
func (s *Server) ListenAndServe() error {
	log.Printf("listening on %s", s.server.Addr)
	if err := s.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

// Stop waits for open requests to finish, until ctx ends
func (s *Server) Stop(ctx context.Context) {
	if err := s.server.Shutdown(ctx); err != nil {
		log.Printf("shutdown: %v", err)
	}
}
