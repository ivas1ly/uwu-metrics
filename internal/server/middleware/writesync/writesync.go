package writesync

import (
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"

	"github.com/ivas1ly/uwu-metrics/internal/server/storage/persistent/file"
)

func New(log *zap.Logger, storage file.PersistentStorage) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		log = log.With(zap.String("middleware", "write sync"))

		log.Info("added write sync middleware")

		syncFn := func(w http.ResponseWriter, r *http.Request) {
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

			next.ServeHTTP(ww, r)

			header := ww.Header().Get("Content-Type")
			if ww.Status() == http.StatusOK && r.Method == http.MethodPost && strings.Contains(header, "text/plain") ||
				strings.Contains(header, "application/json") {
				if err := storage.Save(); err != nil {
					log.Error("can't save metrics", zap.Error(err))
				} else {
					log.Info("all metrics saved successfully")
				}
			}
		}

		return http.HandlerFunc(syncFn)
	}
}
