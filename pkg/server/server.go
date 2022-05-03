package server

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"sort"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-http-utils/headers"
	"github.com/ldez/mimetype"
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
		r.With(withContentType(mimetype.ApplicationJSON)).Post("/register", srv.register)
		r.With(withContentType(mimetype.ApplicationJSON)).Post("/login", srv.login)
		r.Group(func(r chi.Router) {
			r.Use(srv.jwtAuth.CheckAuthentication)
			r.With(withContentType(mimetype.TextPlain)).Post("/orders", srv.createOrder)
			r.Get("/orders", srv.getOrder)

			r.Route("/balance", func(r chi.Router) {
				r.Get("/", srv.currentBalance)
				r.Post("/withdraw", srv.withdraw)
			})
		})

	})

	srv.Get("/ping", srv.Ping())

	return srv
}

func (s *Server) register(w http.ResponseWriter, r *http.Request) {
	log := logger2.FromContext(r.Context())
	body, err := ioutil.ReadAll(r.Body)
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

func (s *Server) Ping() http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		log := logger2.FromContext(request.Context())
		log.Info("Blah")
		writer.WriteHeader(http.StatusOK)
		writer.Header().Set(headers.ContentType, "application/text")

		writer.Write([]byte("pong"))
	}
}

func (s *Server) login(w http.ResponseWriter, r *http.Request) {
	log := logger2.FromContext(r.Context())
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Error("Error reading body", zap.Error(err))
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

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Error("Could not read request body", zap.Error(err))
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}

	orderID, err := orderIDFromBytes(body)
	if err != nil {
		log.Error("Could not create orderID", zap.Error(err))
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}

	o := models.NewOrder(orderID, userID)
	if !o.Valid() {
		log.Error("OrderIS is invalid")
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}

	err = s.orderRepo.CreateNewOrder(r.Context(), o)
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

func orderIDFromBytes(order []byte) (int, error) {
	orderID, err := strconv.Atoi(string(order))
	if err != nil {
		return -1, err
	}
	return orderID, nil
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

	w.Header().Set(headers.ContentType, mimetype.ApplicationJSON)
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(o); err != nil {
		log.Error("Error encoding orders", zap.Error(err))
	}
}

func (s *Server) withdraw(w http.ResponseWriter, r *http.Request) {
	log := logger2.FromContext(r.Context())
	userID, err := controllers.UserIDFromContext(r.Context())
	if err != nil {
		log.Error("orders", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Error("Error reading body", zap.Error(err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var withdrawal controllers.Withdrawal
	err = json.Unmarshal(body, &withdrawal)
	if err != nil {
		log.Error("Error decoding withdrawal", zap.Error(err), zap.ByteString("body", body))
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	order, err := withdrawal.ToOrder(userID)
	if err != nil {
		log.Error("Error creating order", zap.Error(err))
		w.WriteHeader(http.StatusBadRequest)
	}

	if !order.Valid() {
		log.Error("Invalid order id")
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}

	err = s.orderRepo.Withdraw(r.Context(), order)
	switch {
	case err == repo.ErrNotEnoughFunds:
		log.Info("Not enough funds")
		w.WriteHeader(http.StatusPaymentRequired)
		return
	case err == nil:
		w.WriteHeader(http.StatusOK)
		return
	default:
		log.Error("Internal error", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (s *Server) currentBalance(w http.ResponseWriter, r *http.Request) {
	log := logger2.FromContext(r.Context())
	userID, err := controllers.UserIDFromContext(r.Context())
	if err != nil {
		log.Error("orders", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	balancer, err := s.orderRepo.CurrentBalance(r.Context(), userID)
	if err != nil {
		log.Error("Internal error", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set(headers.ContentType, mimetype.ApplicationJSON)
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(balancer)
	if err != nil {
		log.Error("Error encoding response", zap.Error(err))
	}
}
