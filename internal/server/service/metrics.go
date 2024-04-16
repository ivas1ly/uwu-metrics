package service

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/ivas1ly/uwu-metrics/internal/server/entity"
)

type MetricsRepository interface {
	UpdateCounter(name string, value int64)
	UpdateGauge(name string, value float64)
	GetCounter(name string) (int64, error)
	GetGauge(name string) (float64, error)
	GetMetrics() entity.Metrics
}

type MetricsService struct {
	metricsRepository MetricsRepository
}

func NewMetricsService(metricsRepository MetricsRepository) *MetricsService {
	return &MetricsService{
		metricsRepository: metricsRepository,
	}
}

func (s MetricsService) UpsertMetric(mType, mName, mValue string) error {
	switch mType {
	case entity.GaugeType:
		value, err := strconv.ParseFloat(mValue, 64)
		if err != nil {
			return fmt.Errorf("%w; %w; test", err, entity.ErrIncorrectMetricValue)
		}
		s.metricsRepository.UpdateGauge(mName, value)
	case entity.CounterType:
		value, err := strconv.ParseInt(mValue, 10, 64)
		if err != nil {
			return fmt.Errorf("%w: %w", err, entity.ErrIncorrectMetricValue)
		}
		s.metricsRepository.UpdateCounter(mName, value)
	default:
		return entity.ErrUnknownMetricType
	}

	return nil
}

func (s *MetricsService) GetMetric(mType, mName string) (*int64, *float64, error) {
	switch mType {
	case entity.CounterType:
		delta, err := s.metricsRepository.GetCounter(mName)
		if err != nil {
			return nil, nil, errors.Join(err, entity.ErrCanNotGetMetricValue)
		}
		return &delta, nil, nil
	case entity.GaugeType:
		value, err := s.metricsRepository.GetGauge(mName)
		if err != nil {
			return nil, nil, errors.Join(err, entity.ErrCanNotGetMetricValue)
		}
		return nil, &value, nil
	}

	return nil, nil, entity.ErrUnknownMetricType
}

func (s *MetricsService) GetAllMetrics() entity.Metrics {
	return s.metricsRepository.GetMetrics()
}

func (s *MetricsService) UpsertTypeMetric(metric *entity.Metric) (*entity.Metric, error) {
	switch metric.MType {
	case entity.GaugeType:
		if metric.Value == nil {
			return nil, entity.ErrEmptyMetricValue
		}
		s.metricsRepository.UpdateGauge(metric.ID, *metric.Value)

		value, err := s.metricsRepository.GetGauge(metric.ID)
		if err != nil {
			return nil, errors.Join(err, entity.ErrCanNotGetMetricValue)
		}

		metric.Value = &value
		metric.Delta = nil

		return metric, nil
	case entity.CounterType:
		if metric.Delta == nil {
			return nil, entity.ErrEmptyMetricValue
		}
		s.metricsRepository.UpdateCounter(metric.ID, *metric.Delta)

		delta, err := s.metricsRepository.GetCounter(metric.ID)
		if err != nil {
			return nil, errors.Join(err, entity.ErrCanNotGetMetricValue)
		}

		metric.Delta = &delta
		metric.Value = nil

		return metric, nil
	}
	return nil, entity.ErrUnknownMetricType
}
