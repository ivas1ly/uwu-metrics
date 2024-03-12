package checkhash

import (
	"bytes"
	"io"
	"net/http"

	"github.com/go-chi/render"
	"go.uber.org/zap"

	"github.com/ivas1ly/uwu-metrics/internal/utils/hash"
)

// New constructs a new SHA256 hash check middleware.
func New(log *zap.Logger, key []byte) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		l := log.With(zap.String("middleware", "check hash"))

		l.Info("added check hash middleware")

		checkHashFn := func(w http.ResponseWriter, r *http.Request) {
			if len(key) == 0 {
				l.Info("key is empty, skip hash check")
				next.ServeHTTP(w, r)
				return
			}

			hashHeader := r.Header.Get("HashSHA256")
			l.Info("hash", zap.String("header", hashHeader))

			if hashHeader == "" {
				l.Info("hash header is empty, skip check")
				next.ServeHTTP(w, r)
				return
			}

			buf, err := io.ReadAll(r.Body)
			if err != nil {
				l.Info("can't read body")

				w.WriteHeader(http.StatusBadRequest)
				render.JSON(w, r, render.M{"message": "can't read body"})
				return
			}

			sign, err := hash.Hash(buf, key)
			if err != nil {
				l.Info("can't get hash sign")

				w.WriteHeader(http.StatusBadRequest)
				render.JSON(w, r, render.M{"message": "can't check hash"})
				return
			}
			l.Info("hash", zap.String("value", sign))

			if sign != hashHeader {
				l.Info("computed hash doesn't match the one provided in the HashSHA256 header")

				w.WriteHeader(http.StatusBadRequest)
				render.JSON(w, r, render.M{"message": "can't check hash"})
				return
			}

			l.Info("hash check OK", zap.String("header", hashHeader), zap.String("sign", sign))

			reader := io.NopCloser(bytes.NewBuffer(buf))
			r.Body = reader

			next.ServeHTTP(w, r)
		}

		return http.HandlerFunc(checkHashFn)
	}
}
