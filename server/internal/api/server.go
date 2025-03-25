package api

import (
	"context"
	"net/http"
	"os"
	"time"
)

type Server struct {
	server *http.Server
}

func NewServer() *Server {
	port := os.Getenv("Port")
	if port == "" {
		port = "8080"
	}

	router := Routes()

	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	return &Server{
		server: srv,
	}
}

// Start runs the HTTP server
func (s *Server) Start() error {
	return s.server.ListenAndServe()
}

// Shutdown gracefully stops the server
func (s *Server) Shutdown() error {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// Attempt to gracefully shut down the server
	return s.server.Shutdown(ctx)

}
