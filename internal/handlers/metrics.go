package handlers

import (
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/ivas1ly/uwu-metrics/internal/metrics"
	"github.com/ivas1ly/uwu-metrics/internal/storage"
)

type MetricsHandler struct {
	storage storage.Storage
	logger  *slog.Logger
}

func NewMetricsHandler(storage storage.Storage, logger *slog.Logger) *MetricsHandler {
	return &MetricsHandler{
		storage: storage,
		logger:  logger,
	}
}

func (h *MetricsHandler) NewMetricsRoutes(mux *http.ServeMux) {
	mux.Handle("/update/", http.StripPrefix("/update/", http.HandlerFunc(h.update)))
}

func (h *MetricsHandler) update(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	slog.Debug("url path", slog.String("url path", r.URL.Path))
	splitedPath := strings.Split(r.URL.Path, "/")
	if len(splitedPath) == 1 || splitedPath[1] == "" {
		slog.Warn("can't get metric name in url", slog.String("path", r.URL.Path))
		http.NotFound(w, r)
		return
	}

	var mType, mName, mValue string
	if len(splitedPath) == 3 {
		mType, mName, mValue = splitedPath[0], splitedPath[1], splitedPath[2]
	}

	if mValue == "" {
		h.logger.Error("incorrect metric value", slog.String("value", mValue))
		http.Error(w, fmt.Sprintf("incorrect metric value: %q", mValue), http.StatusBadRequest)
		return
	}

	metric := metrics.Metric{
		Type:  mType,
		Name:  mName,
		Value: mValue,
	}

	err := h.storage.Update(metric)
	if err != nil {
		h.logger.Error("incorrect metric type or value", slog.String("error", err.Error()))
		http.Error(w, fmt.Sprintf("incorrect metric type or value; recieved type: %q, value: %q", mType, mValue), http.StatusBadRequest)
		return
	}
	h.logger.Info("metric saved")

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
}
