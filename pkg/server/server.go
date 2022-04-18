package server

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"

	"github.com/OmAsana/go-yapraktikum-final/pkg/logger"
)

type Server struct {
	*chi.Mux
	logger *zap.Logger
}

func NewServer(logger *zap.Logger) *Server {
	srv := &Server{
		Mux:    chi.NewMux(),
		logger: logger}
	setupRoutes(srv)
	return srv
}

func setupRoutes(srv *Server) {
	srv.Use(middleware.RequestID)
	srv.Use(middleware.RealIP)
	srv.Use(logger.Logger)
	srv.Use(middleware.Recoverer)

	srv.Get("/ping", srv.Ping())
}

func (s Server) Ping() http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		log := logger.FromContext(request.Context())
		log.Info("Blah")
		writer.WriteHeader(http.StatusOK)
		writer.Header().Set("Content-Type", "application/text")
		writer.Write([]byte("pong"))
		return
	}
}
