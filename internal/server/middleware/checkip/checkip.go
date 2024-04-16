package checkip

import (
	"net"
	"net/http"

	"github.com/go-chi/render"
	"go.uber.org/zap"
)

// New constructs a new middleware to check if an IP address is in a trusted subnet.
func New(log *zap.Logger, trustedSubnet *net.IPNet) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		l := log.With(zap.String("middleware", "check ip address"))

		l.Info("added check ip address middleware")

		checkHashFn := func(w http.ResponseWriter, r *http.Request) {
			realIP := r.Header.Get("X-Real-IP")
			l.Info("IP address", zap.String("X-Real-IP", realIP))

			requestIP := net.ParseIP(realIP)
			if net.ParseIP(realIP) == nil {
				l.Info("X-Real-IP header is empty or invalid, skip check")
				next.ServeHTTP(w, r)
				return
			}

			if !trustedSubnet.Contains(requestIP) {
				l.Warn("ip address is not in trusted subnet")

				w.WriteHeader(http.StatusForbidden)
				render.JSON(w, r, render.M{"message": "can't check hash"})
				return
			}

			l.Info("ip address check OK", zap.String("X-Real-IP", realIP))

			next.ServeHTTP(w, r)
		}

		return http.HandlerFunc(checkHashFn)
	}
}
