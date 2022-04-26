package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"sort"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/theplant/luhn"
	"go.uber.org/zap"

	"github.com/OmAsana/go-yapraktikum-final/pkg/controllers"
	"github.com/OmAsana/go-yapraktikum-final/pkg/jwt"
	logger2 "github.com/OmAsana/go-yapraktikum-final/pkg/logger"
	"github.com/OmAsana/go-yapraktikum-final/pkg/models"
	"github.com/OmAsana/go-yapraktikum-final/pkg/repo"
)

type Server struct {
	*chi.Mux
	logger    *zap.Logger
	userRepo  repo.UserRepository
	orderRepo repo.OrderRepository
	jwtAuth   *jwt.Authentication
}

func NewServer(logger *zap.Logger, userRepo repo.UserRepository, orderRepo repo.OrderRepository, salt string) *Server {
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
		r.Group(func(r chi.Router) {
			r.Use(srv.jwtAuth.CheckAuthentication)
			r.Post("/orders", srv.createOrder)
			r.Get("/orders", srv.getOrder)

		})
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
		log.Error("Error decoding user", zap.Error(err), zap.ByteString("body", body))
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	log.Info("Registering new user", zap.String("user", creds.Login))

	userID, err := s.userRepo.Create(r.Context(), creds.Login, creds.Password)
	if err != nil {
		if errors.Is(err, repo.ErrUserAlreadyExists) {
			log.Error("User already exists", zap.String("user", creds.Login))
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
	w.WriteHeader(http.StatusOK)
}

func validateRequest(r *http.Request) ([]byte, error) {
	if !Contains(r.Header.Values("Content-Type"), "application/json") {
		return nil, fmt.Errorf("wrong content type: %s", r.Header.Values("Content-Type"))
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

func (s *Server) createOrder(w http.ResponseWriter, r *http.Request) {
	log := logger2.FromContext(r.Context())
	userID, err := controllers.UserIDFromContext(r.Context())
	if err != nil {
		log.Error("orders", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if !Contains(r.Header.Values("Content-Type"), "text/plain") {
		log.Error("Wrong content type")
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Error("Error reading body", zap.Error(err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	orderID, err := strconv.Atoi(string(body))
	if err != nil {
		fmt.Println(err)
		log.Error("Error converting body to int", zap.Error(err))
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}
	if !luhn.Valid(orderID) {
		log.Error("Invalid order id", zap.Int("order_id", orderID))
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}

	err = s.orderRepo.CreateNewOrder(r.Context(), models.NewOrder(orderID, userID))
	switch {
	case err == nil:
		w.WriteHeader(http.StatusAccepted)
		return
	case errors.Is(err, repo.ErrOrderAlreadyUploadedByCurrentUser):
		w.WriteHeader(http.StatusOK)
		return
	case errors.Is(err, repo.ErrOrderCreatedByAnotherUser):
		w.WriteHeader(http.StatusConflict)
		return
	case errors.Is(err, repo.ErrInternalError):
		log.Error("Error creating order", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (s *Server) getOrder(w http.ResponseWriter, r *http.Request) {
	log := logger2.FromContext(r.Context())
	userID, err := controllers.UserIDFromContext(r.Context())
	if err != nil {
		log.Error("orders", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	orders, err := s.orderRepo.ListOrders(r.Context(), userID)
	if err != nil {
		log.Error("Could not retrieve orders", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if len(orders) == 0 {
		log.Info("No orders to return")
		w.WriteHeader(http.StatusNoContent)
		return
	}

	sort.Slice(orders, func(i, j int) bool {
		return orders[i].UploadedAt.Before(orders[j].UploadedAt)
	})

	var o []controllers.Order
	for _, v := range orders {
		o = append(o, controllers.OrderModelToController(*v))
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(o); err != nil {
		log.Error("Error encoding orders", zap.Error(err))
	}
	return
}

func Contains(list []string, value string) bool {
	for _, v := range list {
		return v == value
	}
	return false
}
