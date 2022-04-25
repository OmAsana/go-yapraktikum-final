package jwt

import (
	"net/http"
	"time"

	jwtgo "github.com/dgrijalva/jwt-go"
)

type Claims struct {
	UserID int `json:"user_id"`
	jwtgo.StandardClaims
}

type Authentication struct {
	salt string
}

func NewAuthentication(salt string) *Authentication {
	return &Authentication{salt: salt}
}

func (a *Authentication) CreateClaim(userID int) (*http.Cookie, error) {
	expirationTime := time.Now().Add(5 * time.Minute)
	claims := &Claims{
		UserID: userID,
		StandardClaims: jwtgo.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}

	token := jwtgo.NewWithClaims(jwtgo.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(a.salt))
	if err != nil {
		return nil, err
	}

	return &http.Cookie{
		Name:    "token",
		Value:   tokenString,
		Expires: expirationTime,
	}, nil
}
