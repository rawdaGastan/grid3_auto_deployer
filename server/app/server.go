// Package app for c4s backend app
package app

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rs/zerolog/log"
)

var substrateBlockDiffInSeconds = 6

// Server struct holds port of server
type server struct {
	host string
	port string
}

// NewServer create new server with all configurations
func newServer(host, port string) *server {
	return &server{host, port}
}

// Start starts the server
func (s *server) start() (err error) {
	log.Info().Msgf("Server is listening on %s%s", s.host, s.port)

	srv := &http.Server{
		Addr: s.port,
	}

	go func() {
		if err := srv.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			log.Fatal().Err(err).Msg("HTTP server error")
		}
		log.Info().Msg("Stopped serving new connections")
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	shutdownCtx, shutdownRelease := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownRelease()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatal().Err(err).Msg("HTTP shutdown error")
	}
	log.Info().Msg("Graceful shutdown complete")

	return nil
}
