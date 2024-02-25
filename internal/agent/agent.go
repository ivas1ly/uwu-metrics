package agent

import (
	"context"
	"net/http"
	_ "net/http/pprof" //nolint:gosec // exposed on a separate port that should be unavailable
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"

	"github.com/ivas1ly/uwu-metrics/internal/agent/metrics"
	"github.com/ivas1ly/uwu-metrics/internal/lib/logger"
)

func Run(cfg Config) {
	log := logger.New(defaultLogLevel, logger.NewDefaultLoggerConfig()).
		With(zap.String("app", "agent"))

	ctx := context.Background()
	withCancel, cancel := context.WithCancel(ctx)
	defer cancel()

	endpoint := url.URL{
		Scheme: "http",
		Host:   cfg.EndpointHost,
		Path:   "/updates/",
	}

	ms := &metrics.Metrics{}

	go func() {
		log.Info("start pprof server")
		//nolint:gosec // use the default configuration for pprof
		if err := http.ListenAndServe(defaultPprofAddr, nil); err != nil {
			log.Fatal("pprof server", zap.Error(err))
		}
	}()

	client := &Client{
		URL:     endpoint.String(),
		Metrics: ms,
		Logger:  log,
		Key:     []byte(cfg.Key),
	}
	log.Info("agent started", zap.String("server endpoint", cfg.EndpointHost),
		zap.Duration("pollInterval", cfg.PollInterval), zap.Duration("reportInterval", cfg.ReportInterval))

	notifyCtx, stop := signal.NotifyContext(withCancel, os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)
	defer stop()

	go metrics.RunMetricsUpdate(notifyCtx, ms, cfg.PollInterval, log)
	go metrics.RunPsutilMetricsUpdate(notifyCtx, ms, cfg.PollInterval, log)

	go runReportSend(notifyCtx, client, cfg.ReportInterval, log)

	// block until signal is received
	<-notifyCtx.Done()

	log.Info("app got os signal", zap.String("signal", notifyCtx.Err().Error()))
	log.Info("shutting down...")
	// stop receiving signal notifications as soon as possible.
	stop()

	log.Info("shutdown successfully")
}

func runReportSend(ctx context.Context, client *Client, reportInterval time.Duration, log *zap.Logger) {
	reportTicker := time.NewTicker(reportInterval)
	defer reportTicker.Stop()

	log.Info("start update metrics job", zap.Duration("interval", reportInterval))

	for {
		select {
		case <-ctx.Done():
			log.Info("received done context")
			return
		case rt := <-reportTicker.C:
			log.Info("[report] metrics sent to server", zap.Time("sent at", rt))
			err := client.SendReport()
			if err != nil {
				log.Info("[report] failed to send metrics to server")
			}
		}
	}
}
