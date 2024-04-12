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
	"github.com/ivas1ly/uwu-metrics/web"
)

type MetricsService interface {
	UpsertMetric(mType, mName, mValue string) error
	GetMetric(mType, mName string) (*int64, *float64, error)
	GetAllMetrics() entity.Metrics
	UpsertTypeMetric(metric *entity.Metric) (*entity.Metric, error)
}

type metricsHandler struct {
	log            *zap.Logger
	metricsService MetricsService
}

// NewRoutes adds HTTP endpoints to work with metrics.
func NewRoutes(router *chi.Mux, metricsService MetricsService, log *zap.Logger) {
	h := &metricsHandler{
		metricsService: metricsService,
		log:            log.With(zap.String("handler", "metrics")),
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

// updateURL adds the metric specified in the URL to the storage.
func (h *metricsHandler) updateURL(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")

	mType := strings.ToLower(chi.URLParam(r, "type"))
	if mType == "" {
		h.log.Info("can't get metric type in url", zap.String("path", r.URL.Path))
		http.NotFound(w, r)
		return
	}

	mName := chi.URLParam(r, "name")
	if mName == "" {
		h.log.Info("can't get metric name in url", zap.String("path", r.URL.Path))
		http.NotFound(w, r)
		return
	}

	mValue := chi.URLParam(r, "value")

	err := h.metricsService.UpsertMetric(mType, mName, mValue)
	if errors.Is(err, entity.ErrIncorrectMetricValue) {
		h.log.Info(entity.ErrIncorrectMetricValue.Error(), zap.Error(err))
		http.Error(w, fmt.Sprintf("%s %q", err, mValue), http.StatusBadRequest)
		return
	}
	if errors.Is(err, entity.ErrUnknownMetricType) {
		h.log.Info(entity.ErrUnknownMetricType.Error(), zap.String("type", mType))
		http.Error(w, fmt.Sprintf("%s %q", entity.ErrUnknownMetricType.Error(), mType), http.StatusBadRequest)
		return
	}

	h.log.Info("metric saved", zap.String("type", mType), zap.String("name", mName), zap.String("value", mValue))
	h.log.Debug("in storage", zap.String("metrics", fmt.Sprintf("%+v", h.metricsService.GetAllMetrics())))

	w.WriteHeader(http.StatusOK)
}

// valueURL gets the metric from the storage at the specified name in the URL.
func (h *metricsHandler) valueURL(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")

	mType := strings.ToLower(chi.URLParam(r, "type"))
	if mType == "" {
		h.log.Info("can't get metric type in url", zap.String("path", r.URL.Path))
		http.NotFound(w, r)
		return
	}

	mName := chi.URLParam(r, "name")
	if mName == "" {
		h.log.Info("can't get metric name in url", zap.String("path", r.URL.Path))
		http.NotFound(w, r)
		return
	}

	delta, value, err := h.metricsService.GetMetric(mType, mName)
	if errors.Is(err, entity.ErrCanNotGetMetricValue) {
		h.log.Info(entity.ErrCanNotGetMetricValue.Error(), zap.Error(err))
		http.NotFound(w, r)
		return
	}
	if errors.Is(err, entity.ErrUnknownMetricType) {
		h.log.Info(entity.ErrUnknownMetricType.Error(), zap.String("type", mType))
		http.NotFound(w, r)
		return
	}

	switch mType {
	case entity.GaugeType:
		_, _ = w.Write([]byte(strconv.FormatFloat(*value, 'f', -1, 64)))
	case entity.CounterType:
		_, _ = w.Write([]byte(strconv.FormatInt(*delta, 10)))
	}
}

// webpage shows a static HTML page with the current metrics.
func (h *metricsHandler) webpage(w http.ResponseWriter, _ *http.Request) {
	t, err := template.ParseFS(&web.Templates, "templates/*.gohtml")
	if err != nil {
		h.log.Info("can't parse template from fs", zap.String("error", err.Error()))
		http.Error(w, "", http.StatusInternalServerError)
		return
	}

	metrics := h.metricsService.GetAllMetrics()

	viewMap := template.FuncMap{
		"now":     time.Now().Format(time.RFC850),
		"Gauge":   metrics.Gauge,
		"Counter": metrics.Counter,
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	err = t.ExecuteTemplate(w, "index.gohtml", viewMap)
	if err != nil {
		h.log.Info("can't execute template", zap.String("error", err.Error()))
		http.Error(w, "", http.StatusInternalServerError)
		return
	}
}

// MetricReqRes structure for unmarshaling metrics from the request body.
type MetricReqRes struct {
	Delta *int64   `json:"delta,omitempty"`
	Value *float64 `json:"value,omitempty"`
	ID    string   `json:"id"`
	MType string   `json:"type"`
}

// updateJSON adds the metric specified in the request body to the storage.
func (h *metricsHandler) updateJSON(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var request MetricReqRes

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

	upserted, err := h.metricsService.UpsertTypeMetric(&entity.Metric{
		Delta: request.Delta,
		Value: request.Value,
		ID:    request.ID,
		MType: request.MType,
	})
	if errors.Is(err, entity.ErrEmptyMetricValue) {
		h.log.Info(entity.ErrEmptyMetricValue.Error(), zap.String("type", request.MType))
		w.WriteHeader(http.StatusBadRequest)
		render.JSON(w, r, render.M{"message": fmt.Sprintf("%s %q",
			entity.ErrEmptyMetricValue.Error(), request.MType)})
		return
	}
	if errors.Is(err, entity.ErrUnknownMetricType) {
		h.log.Info(entity.ErrUnknownMetricType.Error(), zap.String("type", request.MType))
		w.WriteHeader(http.StatusBadRequest)
		render.JSON(w, r, render.M{"message": fmt.Sprintf("%s %q",
			entity.ErrUnknownMetricType.Error(), request.MType)})
		return
	}
	if errors.Is(err, entity.ErrCanNotGetMetricValue) {
		h.log.Info("can't get updated value", zap.String("type", request.MType), zap.String("name", request.ID))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	render.JSON(w, r, &MetricReqRes{
		Delta: upserted.Delta,
		Value: upserted.Value,
		ID:    upserted.ID,
		MType: upserted.MType,
	})

	h.log.Info("metric saved",
		zap.String("type", request.MType),
		zap.String("name", request.ID))
	h.log.Debug("in storage", zap.String("metrics", fmt.Sprintf("%+v", h.metricsService.GetAllMetrics())))
}

// updatesJSON adds the array of metrics specified in the body of the request to the storage.
func (h *metricsHandler) updatesJSON(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var request []MetricReqRes
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

		_, err := h.metricsService.UpsertTypeMetric(&entity.Metric{
			Delta: metric.Delta,
			Value: metric.Value,
			ID:    metric.ID,
			MType: metric.MType,
		})
		if errors.Is(err, entity.ErrEmptyMetricValue) {
			h.log.Info(entity.ErrEmptyMetricValue.Error(), zap.String("type", metric.MType))
			w.WriteHeader(http.StatusBadRequest)
			render.JSON(w, r, render.M{"message": fmt.Sprintf("%s %q",
				entity.ErrEmptyMetricValue.Error(), metric.MType)})
			return
		}
		if errors.Is(err, entity.ErrUnknownMetricType) {
			h.log.Info(entity.ErrUnknownMetricType.Error(), zap.String("type", metric.MType))
			w.WriteHeader(http.StatusBadRequest)
			render.JSON(w, r, render.M{"message": fmt.Sprintf("%s %q",
				entity.ErrUnknownMetricType.Error(), metric.MType)})
			return
		}
		if errors.Is(err, entity.ErrCanNotGetMetricValue) {
			h.log.Info("can't get updated value", zap.String("type", metric.MType), zap.String("name", metric.ID))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	w.WriteHeader(http.StatusOK)
}

// valueJSON gets the metric from the storage by the name and type of
// metric specified in the JSON request body.
func (h *metricsHandler) valueJSON(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var request MetricReqRes

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

	delta, value, err := h.metricsService.GetMetric(request.MType, request.ID)
	if errors.Is(err, entity.ErrCanNotGetMetricValue) {
		h.log.Info(entity.ErrCanNotGetMetricValue.Error(), zap.Error(err))
		fmt.Println(err)
		w.WriteHeader(http.StatusNotFound)
		render.JSON(w, r, render.M{"message": err.Error()})
		return
	}
	if errors.Is(err, entity.ErrUnknownMetricType) {
		h.log.Info(entity.ErrUnknownMetricType.Error(), zap.String("type", request.MType))
		w.WriteHeader(http.StatusBadRequest)
		render.JSON(w, r, render.M{"message": fmt.Sprintf("%s %q",
			entity.ErrUnknownMetricType.Error(), request.MType)})
		return
	}

	switch request.MType {
	case entity.GaugeType:
		render.JSON(w, r, MetricReqRes{
			Delta: nil,
			Value: value,
			ID:    request.ID,
			MType: request.MType,
		})
	case entity.CounterType:
		render.JSON(w, r, MetricReqRes{
			Delta: delta,
			Value: nil,
			ID:    request.ID,
			MType: request.MType,
		})
	}
}

// checkRequestFields method for simple validation of query values.
func checkRequestFields(request MetricReqRes) ([]string, bool) {
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
