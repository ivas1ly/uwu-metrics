package writesync

import (
	"context"
	"time"

	"github.com/ivas1ly/uwu-metrics/internal/server/storage/persistent"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func NewInterceptor(storage persistent.Storage, log *zap.Logger) grpc.UnaryServerInterceptor {
	l := log.With(zap.String("unary interceptor", "write sync"))

	l.Info("added write sync unary interceptor")

	syncFn := func(ctx context.Context, req any, _ *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler) (any, error) {
		resp, err := handler(ctx, req)

		if status.Code(err) == codes.OK {
			for _, interval := range retryIntervals {
				errSave := storage.Save(ctx)
				if errSave != nil {
					l.Info("can't save metrics, trying to save metrics again", zap.Error(errSave),
						zap.Duration("with interval", interval))
					time.Sleep(interval)
				} else {
					l.Info("all metrics saved successfully")
					break
				}
			}
		}

		return resp, err
	}

	return syncFn
}
