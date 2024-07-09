package grpc

import (
	"context"
	"net"
	"time"

	"github.com/ivas1ly/uwu-metrics/internal/agent/metrics"
	pb "github.com/ivas1ly/uwu-metrics/pkg/api/metrics"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/encoding/gzip" // Install the gzip compressor
	"google.golang.org/grpc/status"
)

const (
	defaultPayloadCap    = 40
	defaultClientTimeout = 10 * time.Second
)

// retryPolicy - client config https://github.com/grpc/grpc-go/tree/master/examples/features/retry
//
// Config - https://github.com/grpc/proposal/blob/master/A6-client-retries.md
var retryPolicy = `{
            "methodConfig": [{
				"name": [{"service": "metrics.MetricsService"}],
                "waitForReady": true,

                "retryPolicy": {
                    "MaxAttempts": 3,
                    "InitialBackoff": "1s",
                    "MaxBackoff": "3s",
                    "BackoffMultiplier": 2.0,
                    "RetryableStatusCodes": ["UNAVAILABLE"]
                }
            }]
        }`

type Client interface {
	SendReport() error
}

type gRPCClient struct {
	Metrics  *metrics.Metrics
	Logger   *zap.Logger
	LocalIP  *net.IP
	Endpoint string
}

func NewClient(metrics *metrics.Metrics, localIP *net.IP, endpoint string, logger *zap.Logger) Client {
	return &gRPCClient{
		Metrics:  metrics,
		Logger:   logger,
		LocalIP:  localIP,
		Endpoint: endpoint,
	}
}

func (c *gRPCClient) SendReport() error {
	conn, err := grpc.Dial(c.Endpoint,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultServiceConfig(retryPolicy),
	)
	if err != nil {
		c.Logger.Info("unable to connect to server", zap.Error(err))
		return err
	}
	defer conn.Close()

	payload := make([]*pb.Metric, 0, defaultPayloadCap)

	for key, value := range c.Metrics.PrepareGaugeReport() {
		val := value

		mp := &pb.Metric{
			Id:    key,
			Mtype: metrics.GaugeType,
			Delta: 0,
			Value: val,
		}

		payload = append(payload, mp)
	}

	for key, value := range c.Metrics.PrepareCounterReport() {
		val := value
		mp := &pb.Metric{
			Id:    key,
			Mtype: metrics.CounterType,
			Delta: val,
			Value: 0,
		}

		payload = append(payload, mp)
	}

	client := pb.NewMetricsServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), defaultClientTimeout)
	defer cancel()

	_, err = client.Updates(ctx,
		&pb.MetricsRequest{Metrics: payload},
		grpc.UseCompressor(gzip.Name),
	)
	if err != nil {
		c.Logger.Info("can't send gRPC message",
			zap.String("code", status.Code(err).String()),
			zap.Error(err),
		)
		return err
	}

	return nil
}
