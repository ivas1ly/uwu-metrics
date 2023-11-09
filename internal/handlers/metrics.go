package handlers

import (
	"fmt"
	"html/template"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/ivas1ly/uwu-metrics/internal/entity"
	"github.com/ivas1ly/uwu-metrics/internal/storage"
	"github.com/ivas1ly/uwu-metrics/web"
)

type metricsHandler struct {
	storage storage.Storage
	logger  *slog.Logger
}

func NewRoutes(router *chi.Mux, storage storage.Storage, logger *slog.Logger) {
	h := &metricsHandler{
		storage: storage,
		logger:  logger.With(slog.String("handler", "metrics")),
	}

	router.Post("/update/{type}/{name}/{value}", h.update)
	router.Get("/value/{type}/{name}", h.value)
	router.Get("/", h.webpage)
}

func (h *metricsHandler) update(w http.ResponseWriter, r *http.Request) {
	mType := strings.ToLower(chi.URLParam(r, "type"))
	if mType == "" {
		h.logger.Warn("can't get metric type in url", slog.String("path", r.URL.Path))
		http.NotFound(w, r)
		return
	}

	mName := chi.URLParam(r, "name")
	if mName == "" {
		h.logger.Warn("can't get metric name in url", slog.String("path", r.URL.Path))
		http.NotFound(w, r)
		return
	}

	mValue := chi.URLParam(r, "value")
	if mValue == "" {
		h.logger.Error("can't get metric value", slog.String("value", mValue))
		http.Error(w, fmt.Sprintf("can't get metric value %q", mValue), http.StatusBadRequest)
		return
	}

	metric := entity.Metric{
		Type:  mType,
		Name:  mName,
		Value: mValue,
	}

	err := h.storage.Update(metric)
	if err != nil {
		h.logger.Error("incorrect metric type or value", slog.String("error", err.Error()))
		http.Error(w, fmt.Sprintf("incorrect metric type or value; "+
			"recieved type: %q, value: %q", mType, mValue), http.StatusBadRequest)
		return
	}
	h.logger.Info("metric saved", slog.String("metric", fmt.Sprintf("%+v", metric)))
	h.logger.Debug("in storage", slog.String("metrics", fmt.Sprintf("%+v", h.storage.GetMetrics())))

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
}

func (h *metricsHandler) value(w http.ResponseWriter, r *http.Request) {
	mType := strings.ToLower(chi.URLParam(r, "type"))
	if mType == "" {
		h.logger.Warn("can't get metric type in url", slog.String("path", r.URL.Path))
		http.NotFound(w, r)
		return
	}

	mName := chi.URLParam(r, "name")
	if mName == "" {
		h.logger.Warn("can't get metric name in url", slog.String("path", r.URL.Path))
		http.NotFound(w, r)
		return
	}

	switch mType {
	case "counter":
		value, err := h.storage.GetCounter(mName)
		if err != nil {
			h.logger.Error("can't get counter value", slog.String("error", err.Error()))
			http.NotFound(w, r)
			return
		}
		_, err = w.Write([]byte(strconv.FormatInt(value, 10)))
		if err != nil {
			h.logger.Error("can't format counter value", slog.String("error", err.Error()))
			http.Error(w, "", http.StatusInternalServerError)
			return
		}
	case "gauge":
		value, err := h.storage.GetGauge(mName)
		if err != nil {
			h.logger.Error("can't get gauge value", slog.String("error", err.Error()))
			http.NotFound(w, r)
			return
		}
		_, err = w.Write([]byte(strconv.FormatFloat(value, 'f', -1, 64)))
		if err != nil {
			h.logger.Error("can't format gauge value", slog.String("error", err.Error()))
			http.Error(w, "", http.StatusInternalServerError)
			return
		}
	default:
		h.logger.Error("unknown metric type", slog.String("type", mType))
		http.NotFound(w, r)
		return
	}
}

func (h *metricsHandler) webpage(w http.ResponseWriter, _ *http.Request) {
	t, err := template.ParseFS(&web.Templates, "templates/*.gohtml")
	if err != nil {
		h.logger.Error("can't parse template from fs", slog.String("error", err.Error()))
		http.Error(w, "", http.StatusInternalServerError)
		return
	}

	m := h.storage.GetMetrics()

	viewMap := template.FuncMap{
		"now":     time.Now().Format(time.RFC850),
		"Gauge":   m.Gauge,
		"Counter": m.Counter,
	}

	err = t.ExecuteTemplate(w, "index.gohtml", viewMap)
	if err != nil {
		h.logger.Error("can't execute template", slog.String("error", err.Error()))
		http.Error(w, "", http.StatusInternalServerError)
		return
	}
}
