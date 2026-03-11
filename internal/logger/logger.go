package logger

import (
	"go.uber.org/zap"
	"net/http"
	"time"
)

func Initialize(level string) (*zap.Logger, error) {
	lvl, err := zap.ParseAtomicLevel(level)
	if err != nil {
		return nil, err
	}

	cfg := zap.NewProductionConfig()
	cfg.Level = lvl
	zl, err := cfg.Build()
	if err != nil {
		return nil, err
	}
	return zl, nil
}

type (
	responseData struct {
		status int
		size   int
	}

	responseWriter struct {
		http.ResponseWriter
		data *responseData
	}
)

func (w *responseWriter) WriteHeader(status int) {
	w.ResponseWriter.WriteHeader(status)
	w.data.status = status
}

func (w *responseWriter) Write(b []byte) (int, error) {
	size, err := w.ResponseWriter.Write(b)
	w.data.size += size
	return size, err
}

func WithLogger(log *zap.Logger) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			uri := r.RequestURI
			method := r.Method

			rd := &responseData{
				status: 0,
				size:   0,
			}

			rw := responseWriter{
				ResponseWriter: w,
				data:           rd,
			}

			h.ServeHTTP(&rw, r)

			duration := time.Since(start)

			log.Info("incoming HTTP request",
				zap.String("method", method),
				zap.String("uri", uri),
				zap.Int("status", rd.status),
				zap.Duration("duration", duration),
				zap.Int("size", rd.size),
			)
		})
	}
}

func Err(err error) zap.Field {
	return zap.Error(err)
}

func Field(name string, field any) zap.Field {
	return zap.Any(name, field)
}
