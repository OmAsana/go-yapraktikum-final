package server

import (
	"fmt"
	"net/http"

	"github.com/go-http-utils/headers"

	"github.com/OmAsana/go-yapraktikum-final/pkg/logger"
)

func withContentType(mimeType string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			log := logger.FromContext(r.Context())
			ct := r.Header.Values(headers.ContentType)
			if !Contains(ct, mimeType) {
				log.Error(fmt.Sprintf("Wrong content type. Want: %s. Got: %s", mimeType, ct))
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func Contains(list []string, value string) bool {
	for _, v := range list {
		if v == value {
			return true
		}
	}
	return false
}
