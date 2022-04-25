package logger

import (
	"context"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type ctxKeyType string

var ctxKey ctxKeyType = "loggerCtx"
var atomicLevel = zap.NewAtomicLevelAt(zapcore.InfoLevel)

var logger = zap.NewNop()

func Logger(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

		t1 := time.Now()
		reqID := middleware.GetReqID(r.Context())
		defer func() {
			logger.Info(
				"Served",
				zap.String("proto", r.Proto),
				zap.String("path", r.URL.Path),
				zap.Duration("took", time.Since(t1)),
				zap.Int("status", ww.Status()),
				zap.Int("size", ww.BytesWritten()),
				zap.String("reqId", reqID),
			)
		}()

		next.ServeHTTP(ww, r)
	}
	return http.HandlerFunc(fn)
}

func FromContext(ctx context.Context) *zap.Logger {
	reqID := middleware.GetReqID(ctx)
	return logger.With(zap.String("reqId", reqID))
}

func SetLogLevel(level string) error {
	var lvl zapcore.Level
	if err := lvl.UnmarshalText([]byte(level)); err != nil {
		return err
	}
	logger.Info("Setting log level to: " + level)
	atomicLevel.SetLevel(lvl)
	return nil
}

func NewLogger() *zap.Logger {
	config := zap.NewProductionConfig()
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	l, err := config.Build()
	if err != nil {
		panic(err)
	}
	logger = l

	return logger
}

func NewNoop() *zap.Logger {
	return zap.NewNop()
}

func NewDevLogger() (*zap.Logger, error) {
	return zap.NewDevelopment()
}
