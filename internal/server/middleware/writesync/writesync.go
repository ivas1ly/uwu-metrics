package writesync

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"

	"github.com/ivas1ly/uwu-metrics/internal/server/storage/persistent"
)

var (
	retryIntervals = []time.Duration{1 * time.Second, 3 * time.Second, 5 * time.Second}
)

func New(ctx context.Context, log *zap.Logger, storage persistent.Storage) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		l := log.With(zap.String("middleware", "write sync"))

		l.Info("added write sync middleware")

		syncFn := func(w http.ResponseWriter, r *http.Request) {
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

			next.ServeHTTP(ww, r)

			header := ww.Header().Get("Content-Type")
			//nolint:whitespace //necessary leading newline
			if ww.Status() == http.StatusOK &&
				r.Method == http.MethodPost &&
				strings.Contains(header, "text/plain") ||
				strings.Contains(header, "application/json") {

				for _, interval := range retryIntervals {
					err := storage.Save(ctx)
					if err != nil {
						l.Info("can't save metrics, trying to save metrics again", zap.Error(err),
							zap.Duration("with interval", interval))
						time.Sleep(interval)
					} else {
						l.Info("all metrics saved successfully")
						break
					}
				}
			}
		}

		return http.HandlerFunc(syncFn)
	}
}
