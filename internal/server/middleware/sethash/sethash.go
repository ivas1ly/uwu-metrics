package sethash

import (
	"bytes"
	"io"
	"net/http"

	"github.com/go-chi/render"
	"go.uber.org/zap"

	"github.com/ivas1ly/uwu-metrics/internal/lib/hash"
)

func New(log *zap.Logger, key []byte) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		l := log.With(zap.String("middleware", "set hash"))

		l.Info("added set hash middleware")

		setHashFn := func(w http.ResponseWriter, r *http.Request) {
			if len(key) == 0 {
				l.Info("key is empty, skip hash computation")
				next.ServeHTTP(w, r)
				return
			}

			buf, err := io.ReadAll(r.Body)
			if err != nil {
				l.Info("can't read body")

				w.WriteHeader(http.StatusInternalServerError)
				render.JSON(w, r, render.M{"message": "can't read body"})
				return
			}

			sign, err := hash.New(buf, key)
			if err != nil {
				l.Info("can't get hash sign")

				w.WriteHeader(http.StatusInternalServerError)
				render.JSON(w, r, render.M{"message": "can't set hash"})
				return
			}
			l.Info("hash", zap.String("value", sign))

			w.Header().Set("HashSHA256", sign)

			l.Info("hash added to the response header HashSHA256", zap.String("hash", sign))

			reader := io.NopCloser(bytes.NewBuffer(buf))
			r.Body = reader

			next.ServeHTTP(w, r)
		}

		return http.HandlerFunc(setHashFn)
	}
}
