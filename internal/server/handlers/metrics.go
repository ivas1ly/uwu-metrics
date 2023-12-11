package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"go.uber.org/zap"

	"github.com/ivas1ly/uwu-metrics/internal/server/entity"
	"github.com/ivas1ly/uwu-metrics/internal/server/storage/memory"
	"github.com/ivas1ly/uwu-metrics/web"
)

const (
	IncorrectMetricValueMsg = "incorrect metric value"
	UnknownMetricTypeMsg    = "unknown metric type"
	EmptyMetricValueMsg     = "empty metric value"
)

type metricsHandler struct {
	storage memory.Storage
	logger  *zap.Logger
}

func NewRoutes(router *chi.Mux, storage memory.Storage, logger *zap.Logger) {
	h := &metricsHandler{
		storage: storage,
		logger:  logger.With(zap.String("handler", "metrics")),
	}

	router.Get("/", h.webpage)
	router.Route("/update", func(r chi.Router) {
		r.Post("/", h.updateJSON)
		r.Post("/{type}/{name}/{value}", h.updateURL)
	})
	router.Route("/value", func(r chi.Router) {
		r.Post("/", h.valueJSON)
		r.Get("/{type}/{name}", h.valueURL)
	})
	router.Route("/updates", func(r chi.Router) {
		r.Post("/", h.updatesJSON)
	})
}

func (h *metricsHandler) updateURL(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")

	mType := strings.ToLower(chi.URLParam(r, "type"))
	if mType == "" {
		h.logger.Info("can't get metric type in url", zap.String("path", r.URL.Path))
		http.NotFound(w, r)
		return
	}

	mName := chi.URLParam(r, "name")
	if mName == "" {
		h.logger.Info("can't get metric name in url", zap.String("path", r.URL.Path))
		http.NotFound(w, r)
		return
	}

	mValue := chi.URLParam(r, "value")

	switch mType {
	case entity.GaugeType:
		value, err := strconv.ParseFloat(mValue, 64)
		if err != nil {
			h.logger.Info(IncorrectMetricValueMsg, zap.String("error", err.Error()))
			http.Error(w, fmt.Sprintf("%s %q", IncorrectMetricValueMsg, mValue), http.StatusBadRequest)
			return
		}
		h.storage.UpdateGauge(mName, value)
	case entity.CounterType:
		value, err := strconv.ParseInt(mValue, 10, 64)
		if err != nil {
			h.logger.Info(IncorrectMetricValueMsg, zap.String("error", err.Error()))
			http.Error(w, fmt.Sprintf("%s %q", IncorrectMetricValueMsg, mValue), http.StatusBadRequest)
			return
		}
		h.storage.UpdateCounter(mName, value)
	default:
		h.logger.Info(UnknownMetricTypeMsg, zap.String("type", mType))
		http.Error(w, fmt.Sprintf("%s %q", UnknownMetricTypeMsg, mType), http.StatusBadRequest)
		return
	}

	h.logger.Info("metric saved",
		zap.String("type", mType),
		zap.String("name", mName),
		zap.String("value", mValue))
	h.logger.Debug("in storage", zap.String("metrics", fmt.Sprintf("%+v", h.storage.GetMetrics())))

	w.WriteHeader(http.StatusOK)
}

func (h *metricsHandler) valueURL(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")

	mType := strings.ToLower(chi.URLParam(r, "type"))
	if mType == "" {
		h.logger.Info("can't get metric type in url", zap.String("path", r.URL.Path))
		http.NotFound(w, r)
		return
	}

	mName := chi.URLParam(r, "name")
	if mName == "" {
		h.logger.Info("can't get metric name in url", zap.String("path", r.URL.Path))
		http.NotFound(w, r)
		return
	}

	switch mType {
	case entity.CounterType:
		value, err := h.storage.GetCounter(mName)
		if err != nil {
			h.logger.Info("can't get counter value", zap.String("error", err.Error()))
			http.NotFound(w, r)
			return
		}
		_, err = w.Write([]byte(strconv.FormatInt(value, 10)))
		if err != nil {
			h.logger.Info("can't format counter value", zap.String("error", err.Error()))
			http.Error(w, "", http.StatusInternalServerError)
			return
		}
	case entity.GaugeType:
		value, err := h.storage.GetGauge(mName)
		if err != nil {
			h.logger.Info("can't get gauge value", zap.String("error", err.Error()))
			http.NotFound(w, r)
			return
		}
		_, err = w.Write([]byte(strconv.FormatFloat(value, 'f', -1, 64)))
		if err != nil {
			h.logger.Info("can't format gauge value", zap.String("error", err.Error()))
			http.Error(w, "", http.StatusInternalServerError)
			return
		}
	default:
		h.logger.Info(UnknownMetricTypeMsg, zap.String("type", mType))
		http.NotFound(w, r)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *metricsHandler) webpage(w http.ResponseWriter, _ *http.Request) {
	t, err := template.ParseFS(&web.Templates, "templates/*.gohtml")
	if err != nil {
		h.logger.Info("can't parse template from fs", zap.String("error", err.Error()))
		http.Error(w, "", http.StatusInternalServerError)
		return
	}

	m := h.storage.GetMetrics()

	viewMap := template.FuncMap{
		"now":     time.Now().Format(time.RFC850),
		"Gauge":   m.Gauge,
		"Counter": m.Counter,
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	err = t.ExecuteTemplate(w, "index.gohtml", viewMap)
	if err != nil {
		h.logger.Info("can't execute template", zap.String("error", err.Error()))
		http.Error(w, "", http.StatusInternalServerError)
		return
	}
}

type Metrics struct {
	Delta *int64   `json:"delta,omitempty"`
	Value *float64 `json:"value,omitempty"`
	ID    string   `json:"id"`
	MType string   `json:"type"`
}

func (h *metricsHandler) updateJSON(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var request Metrics

	err := json.NewDecoder(r.Body).Decode(&request)
	if errors.Is(err, io.EOF) {
		w.WriteHeader(http.StatusBadRequest)
		render.JSON(w, r, render.M{"message": "empty request body"})
		return
	}
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		render.JSON(w, r, render.M{"message": "can't parse request body"})
		return
	}

	errMsg, ok := checkRequestFields(request)
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		render.JSON(w, r, render.M{"message": strings.Join(errMsg, ", ")})
		return
	}

	switch request.MType {
	case entity.GaugeType:
		if request.Value == nil {
			h.logger.Info(EmptyMetricValueMsg, zap.String("type", request.MType))
			w.WriteHeader(http.StatusBadRequest)
			render.JSON(w, r, render.M{"message": fmt.Sprintf("%s %q", EmptyMetricValueMsg, request.MType)})
			return
		}
		h.storage.UpdateGauge(request.ID, *request.Value)

		value, err := h.storage.GetGauge(request.ID)
		if err != nil {
			h.logger.Info("can't get updated value", zap.String("type", request.MType),
				zap.String("name", request.ID))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		render.JSON(w, r, Metrics{
			Delta: nil,
			Value: &value,
			ID:    request.ID,
			MType: request.MType,
		})
	case entity.CounterType:
		if request.Delta == nil {
			h.logger.Info(EmptyMetricValueMsg, zap.String("type", request.MType))
			w.WriteHeader(http.StatusBadRequest)
			render.JSON(w, r, render.M{"message": fmt.Sprintf("%s %q", EmptyMetricValueMsg, request.MType)})
			return
		}
		h.storage.UpdateCounter(request.ID, *request.Delta)

		value, err := h.storage.GetCounter(request.ID)
		if err != nil {
			h.logger.Info("can't get updated value", zap.String("type", request.MType),
				zap.String("name", request.ID))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		render.JSON(w, r, Metrics{
			Delta: &value,
			Value: nil,
			ID:    request.ID,
			MType: request.MType,
		})
	default:
		h.logger.Info(UnknownMetricTypeMsg, zap.String("type", request.MType))
		w.WriteHeader(http.StatusBadRequest)
		render.JSON(w, r, render.M{"message": fmt.Sprintf("%s %q", UnknownMetricTypeMsg, request.MType)})
		return
	}

	h.logger.Info("metric saved",
		zap.String("type", request.MType),
		zap.String("name", request.ID))
	h.logger.Debug("in storage", zap.String("metrics", fmt.Sprintf("%+v", h.storage.GetMetrics())))
}

func (h *metricsHandler) updatesJSON(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var request []Metrics
	err := json.NewDecoder(r.Body).Decode(&request)
	if errors.Is(err, io.EOF) {
		w.WriteHeader(http.StatusBadRequest)
		render.JSON(w, r, render.M{"message": "empty request body"})
		return
	}
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		render.JSON(w, r, render.M{"message": "can't parse request body"})
		return
	}

	for _, metric := range request {
		errMsg, ok := checkRequestFields(metric)
		if !ok {
			w.WriteHeader(http.StatusBadRequest)
			render.JSON(w, r, render.M{"message": strings.Join(errMsg, ", ")})
			return
		}

		switch metric.MType {
		case entity.GaugeType:
			if metric.Value == nil {
				h.logger.Info(EmptyMetricValueMsg, zap.String("type", metric.MType))
				w.WriteHeader(http.StatusBadRequest)
				render.JSON(w, r, render.M{"message": fmt.Sprintf("%s %q", EmptyMetricValueMsg, metric.MType)})
				return
			}

			h.storage.UpdateGauge(metric.ID, *metric.Value)
		case entity.CounterType:
			if metric.Delta == nil {
				h.logger.Info(EmptyMetricValueMsg, zap.String("type", metric.MType))
				w.WriteHeader(http.StatusBadRequest)
				render.JSON(w, r, render.M{"message": fmt.Sprintf("%s %q", EmptyMetricValueMsg, metric.MType)})
				return
			}

			h.storage.UpdateCounter(metric.ID, *metric.Delta)
		default:
			h.logger.Info(UnknownMetricTypeMsg, zap.String("type", metric.MType))
			w.WriteHeader(http.StatusBadRequest)
			render.JSON(w, r, render.M{"message": fmt.Sprintf("%s %q", UnknownMetricTypeMsg, metric.MType)})
			return
		}
	}

	w.WriteHeader(http.StatusOK)
}

func (h *metricsHandler) valueJSON(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var request Metrics

	err := json.NewDecoder(r.Body).Decode(&request)
	if errors.Is(err, io.EOF) {
		w.WriteHeader(http.StatusBadRequest)
		render.JSON(w, r, render.M{"message": "empty request body"})
		return
	}
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		render.JSON(w, r, render.M{"message": "can't parse request body"})
		return
	}

	errMsg, ok := checkRequestFields(request)
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		render.JSON(w, r, render.M{"message": strings.Join(errMsg, ", ")})
		return
	}

	switch request.MType {
	case entity.GaugeType:
		value, err := h.storage.GetGauge(request.ID)
		if err != nil {
			h.logger.Info(err.Error(), zap.String("type", request.MType))
			w.WriteHeader(http.StatusNotFound)
			render.JSON(w, r, render.M{"message": err.Error()})
			return
		}

		render.JSON(w, r, Metrics{
			Delta: nil,
			Value: &value,
			ID:    request.ID,
			MType: request.MType,
		})
	case entity.CounterType:
		value, err := h.storage.GetCounter(request.ID)
		if err != nil {
			h.logger.Info(err.Error(), zap.String("type", request.MType))
			w.WriteHeader(http.StatusNotFound)
			render.JSON(w, r, render.M{"message": err.Error()})
			return
		}

		render.JSON(w, r, Metrics{
			Delta: &value,
			Value: nil,
			ID:    request.ID,
			MType: request.MType,
		})
	default:
		h.logger.Info(UnknownMetricTypeMsg, zap.String("type", request.MType))
		w.WriteHeader(http.StatusBadRequest)
		render.JSON(w, r, render.M{"message": fmt.Sprintf("%s %q", UnknownMetricTypeMsg, request.MType)})
		return
	}
}

func checkRequestFields(request Metrics) ([]string, bool) {
	var errMsg []string

	if strings.TrimSpace(request.MType) == "" {
		errMsg = append(errMsg, fmt.Sprintf("field %q is required", "type"))
	}
	if strings.TrimSpace(request.ID) == "" {
		errMsg = append(errMsg, fmt.Sprintf("field %q is required", "id"))
	}
	if len(errMsg) > 0 {
		return errMsg, false
	}

	return nil, true
}
