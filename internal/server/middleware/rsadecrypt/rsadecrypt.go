package rsadecrypt

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"io"
	"net/http"

	"github.com/go-chi/render"
	"go.uber.org/zap"
)

// New constructs middleware to decrypt body with RSA private key.
func New(log *zap.Logger, key *rsa.PrivateKey) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		l := log.With(zap.String("middleware", "rsa decrypt"))

		l.Info("added rsa decrypt middleware")

		setHashFn := func(w http.ResponseWriter, r *http.Request) {
			if key == nil {
				l.Info("rsa private key is empty, skip body decryption")
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

			decrypted, err := rsa.DecryptPKCS1v15(rand.Reader, key, buf)
			if err != nil {
				l.Info("can't get decrypt body")

				w.WriteHeader(http.StatusInternalServerError)
				render.JSON(w, r, render.M{"message": "can't get decrypt body"})
				return
			}
			l.Info("body decrypted")

			r.Body = io.NopCloser(bytes.NewBuffer(decrypted))

			next.ServeHTTP(w, r)
		}

		return http.HandlerFunc(setHashFn)
	}
}
