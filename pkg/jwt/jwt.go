package jwt

import (
	"context"
	"errors"
	"net/http"
	"time"

	jwtgo "github.com/dgrijalva/jwt-go"
	"go.uber.org/zap"

	"github.com/OmAsana/go-yapraktikum-final/pkg/controllers"
	"github.com/OmAsana/go-yapraktikum-final/pkg/logger"
)

var cookieKey = "token"

type Claims struct {
	UserID int `json:"user_id"`
	jwtgo.StandardClaims
}

type Authentication struct {
	salt []byte
}

func NewAuthentication(salt string) *Authentication {
	return &Authentication{salt: []byte(salt)}
}

func (a *Authentication) CreateClaim(userID int) (*http.Cookie, error) {
	expirationTime := time.Now().Add(10 * time.Hour)
	claims := &Claims{
		UserID: userID,
		StandardClaims: jwtgo.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}

	token := jwtgo.NewWithClaims(jwtgo.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(a.salt)
	if err != nil {
		return nil, err
	}

	return &http.Cookie{
		Name:    cookieKey,
		Value:   tokenString,
		Expires: expirationTime,
	}, nil
}

func (a *Authentication) CheckAuthentication(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log := logger.FromContext(r.Context())
		c, err := r.Cookie(cookieKey)
		if err != nil {
			log.Error("Can not find cookie", zap.Error(err))
			if errors.Is(err, http.ErrNoCookie) {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		tokenStr := c.Value
		claim := &Claims{}

		tkn, err := jwtgo.ParseWithClaims(tokenStr, claim, func(token *jwtgo.Token) (interface{}, error) {
			return a.salt, nil
		})

		if err != nil {

			if !tkn.Valid {
				log.Info("User token is invalid", zap.Error(err))
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			log.Error("Error parsing jwt token", zap.Error(err))
			w.WriteHeader(http.StatusBadRequest)
			return

		}

		if !tkn.Valid {
			log.Info("User token is invalid")
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), controllers.UserCTXKey, claim.UserID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
