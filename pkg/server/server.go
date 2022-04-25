package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"

	"github.com/OmAsana/go-yapraktikum-final/pkg/controllers"
	"github.com/OmAsana/go-yapraktikum-final/pkg/jwt"
	logger2 "github.com/OmAsana/go-yapraktikum-final/pkg/logger"
	"github.com/OmAsana/go-yapraktikum-final/pkg/repo"
)

type Server struct {
	*chi.Mux
	logger    *zap.Logger
	userRepo  repo.User
	orderRepo repo.Order
	jwtAuth   *jwt.Authentication
}

func NewServer(logger *zap.Logger, userRepo repo.User, orderRepo repo.Order, salt string) *Server {
	srv := &Server{
		Mux:       chi.NewMux(),
		logger:    logger,
		userRepo:  userRepo,
		orderRepo: orderRepo,
		jwtAuth:   jwt.NewAuthentication(salt),
	}
	srv.Use(middleware.RequestID)
	srv.Use(middleware.RealIP)
	srv.Use(logger2.Logger)
	srv.Use(middleware.Recoverer)

	srv.Route("/api/user", func(r chi.Router) {
		r.Post("/register", srv.register)
		r.Post("/login", srv.login)
	})

	srv.Get("/ping", srv.Ping())

	return srv
}

func (s *Server) register(w http.ResponseWriter, r *http.Request) {
	log := logger2.FromContext(r.Context())
	body, err := validateRequest(r)
	if err != nil {
		log.Error("Error validating request", zap.Error(err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var creds controllers.Credentials
	err = json.Unmarshal(body, &creds)
	if err != nil {
		log.Error("err decoding user", zap.Error(err), zap.ByteString("body", body))
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	log.Info("Registering new user", zap.String("user", creds.Login))

	userID, err := s.userRepo.Create(r.Context(), creds.Login, creds.Password)
	if err != nil {
		if errors.Is(err, repo.ErrUserAlreadyExists) {
			log.Error("user already exists", zap.String("user", creds.Login))
			w.WriteHeader(http.StatusConflict)
			return
		}
		if errors.Is(err, repo.ErrInternalError) {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	claim, err := s.jwtAuth.CreateClaim(userID)
	if err != nil {
		log.Error("could not create jwt claim", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	http.SetCookie(w, claim)
}

func validateRequest(r *http.Request) ([]byte, error) {
	if !Contains(r.Header.Values("Content-Type"), "application/json") {
		return nil, errors.New(fmt.Sprintf("wrong content type: %s", r.Header.Values("Content-Type")))
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	return body, nil
}

func (s *Server) Ping() http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		log := logger2.FromContext(request.Context())
		log.Info("Blah")
		writer.WriteHeader(http.StatusOK)
		writer.Header().Set("Content-Type", "application/text")
		writer.Write([]byte("pong"))
	}
}

func (s *Server) login(w http.ResponseWriter, r *http.Request) {
	log := logger2.FromContext(r.Context())
	body, err := validateRequest(r)
	if err != nil {
		log.Error("Error validating request", zap.Error(err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var creds controllers.Credentials
	err = json.Unmarshal(body, &creds)
	if err != nil {
		log.Error("Error decoding user", zap.Error(err), zap.ByteString("body", body))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	userID, err := s.userRepo.Authenticate(r.Context(), creds.Login, creds.Password)
	if err != nil {
		log.Error("Error authenticating user", zap.Error(err))
		if errors.Is(err, repo.ErrUserAuthFailed) {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	claim, err := s.jwtAuth.CreateClaim(userID)
	if err != nil {
		log.Error("could not create jwt claim", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	http.SetCookie(w, claim)
	w.WriteHeader(http.StatusOK)
}

func Contains(list []string, value string) bool {
	for _, v := range list {
		return v == value
	}
	return false
}
