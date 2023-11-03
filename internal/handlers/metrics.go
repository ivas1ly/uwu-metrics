package handlers

import (
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"

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

func (h *MetricsHandler) NewMetricsRoutes(router *chi.Mux) {
	router.Post("/update/{type}/{name}/{value}", h.update)
}

func (h *MetricsHandler) update(w http.ResponseWriter, r *http.Request) {
	mType := strings.ToLower(chi.URLParam(r, "type"))
	if mType == "" {
		slog.Warn("can't get metric type in url", slog.String("path", r.URL.Path))
		http.NotFound(w, r)
		return
	}

	mName := chi.URLParam(r, "name")
	if mName == "" {
		slog.Warn("can't get metric name in url", slog.String("path", r.URL.Path))
		http.NotFound(w, r)
		return
	}

	mValue := chi.URLParam(r, "value")
	if mValue == "" {
		h.logger.Error("can't get metric value", slog.String("value", mValue))
		http.Error(w, fmt.Sprintf("can't get metric value %q", mValue), http.StatusBadRequest)
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
	h.logger.Info("metric saved", slog.String("metric", fmt.Sprintf("%+v", metric)))

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
}
