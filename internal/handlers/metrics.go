package handlers

import (
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
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
		http.Error(w, fmt.Sprintf("incorrect metric value %q", mValue), http.StatusBadRequest)
		return
	}

	metric := metrics.Metric{}
	switch mType {
	case "gauge":
		value, err := strconv.ParseFloat(mValue, 64)
		if err != nil {
			h.logger.Error("incorrect metric value", slog.String("value", mValue))
			http.Error(w, fmt.Sprintf("incorrect metric value %q", mValue), http.StatusBadRequest)
			return
		}
		metric.Gauge = value
		metric.Type = mType
	case "counter":
		value, err := strconv.ParseInt(mValue, 10, 64)
		if err != nil {
			h.logger.Error("incorrect metric value", slog.String("value", mValue))
			http.Error(w, fmt.Sprintf("incorrect metric value %q", mValue), http.StatusBadRequest)
			return
		}
		metric.Counter = value
		metric.Type = mType
	default:
		h.logger.Error("incorrect metric type", slog.String("type", mType))
		http.Error(w, fmt.Sprintf("incorrect metric type %q", mType), http.StatusBadRequest)
		return
	}

	h.storage.Update(mName, metric)
	h.logger.Info("metric saved")

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
}
