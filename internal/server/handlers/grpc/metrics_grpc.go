package grpc

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/ivas1ly/uwu-metrics/internal/server/entity"
	pb "github.com/ivas1ly/uwu-metrics/pkg/api/metrics"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type MetricsService interface {
	UpsertMetric(mType, mName, mValue string) error
	GetMetric(mType, mName string) (*int64, *float64, error)
	GetAllMetrics() entity.Metrics
	UpsertTypeMetric(metric *entity.Metric) (*entity.Metric, error)
}

type metricsgRPCHandler struct {
	pb.UnimplementedMetricsServiceServer
	metricsService MetricsService
	log            *zap.Logger
}

func NewRoutes(metricsService MetricsService, log *zap.Logger) *metricsgRPCHandler {
	h := &metricsgRPCHandler{
		metricsService: metricsService,
		log:            log.With(zap.String("gRPC handler", "metrics")),
	}

	return h
}

func (h *metricsgRPCHandler) Updates(ctx context.Context, in *pb.MetricsRequest) (*emptypb.Empty, error) {
	for _, metric := range in.Metrics {
		errMsg, ok := checkRequestFields(metric)
		if !ok {
			return nil, status.Error(codes.InvalidArgument, strings.Join(errMsg, ", "))
		}

		_, err := h.metricsService.UpsertTypeMetric(&entity.Metric{
			Delta: &metric.Delta,
			Value: &metric.Value,
			ID:    metric.Id,
			MType: metric.Mtype,
		})
		if errors.Is(err, entity.ErrEmptyMetricValue) {
			h.log.Info(entity.ErrEmptyMetricValue.Error(), zap.String("type", metric.Mtype))
			return nil, status.Errorf(codes.InvalidArgument, "%s %q",
				entity.ErrEmptyMetricValue.Error(), metric.Mtype)
		}
		if errors.Is(err, entity.ErrUnknownMetricType) {
			h.log.Info(entity.ErrUnknownMetricType.Error(), zap.String("type", metric.Mtype))
			return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("%s %q",
				entity.ErrEmptyMetricValue.Error(), metric.Mtype))
		}
		if errors.Is(err, entity.ErrCanNotGetMetricValue) {
			h.log.Info("can't get updated value", zap.String("type", metric.Mtype), zap.String("name", metric.Id))
			return nil, status.Error(codes.Internal, "")
		}
	}

	return nil, nil
}

// checkRequestFields method for simple validation of query values.
func checkRequestFields(metric *pb.Metric) ([]string, bool) {
	var errMsg []string

	if strings.TrimSpace(metric.Mtype) == "" {
		errMsg = append(errMsg, fmt.Sprintf("field %q is required", "type"))
	}
	if strings.TrimSpace(metric.Id) == "" {
		errMsg = append(errMsg, fmt.Sprintf("field %q is required", "id"))
	}
	if len(errMsg) > 0 {
		return errMsg, false
	}

	return nil, true
}
