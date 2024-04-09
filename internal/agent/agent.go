package agent

import (
	"context"
	"crypto/rsa"
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
	"github.com/ivas1ly/uwu-metrics/internal/utils/rsakeys"
	"github.com/ivas1ly/uwu-metrics/pkg/netutil"
)

// Run starts an agent to collect metrics with the specified configuration.
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

	var publicKey *rsa.PublicKey
	var err error
	if cfg.PublicKeyPath != "" {
		publicKey, err = rsakeys.PublicKey(cfg.PublicKeyPath)
		if err != nil {
			log.Warn("can't get public key from file", zap.Error(err))
		}
		log.Info("public key successfully loaded")
	}

	client := &Client{
		URL:          endpoint.String(),
		Metrics:      ms,
		Logger:       log,
		Key:          []byte(cfg.HashKey),
		RSAPublicKey: publicKey,
		LocalIP:      netutil.GetOutboundIP(),
	}
	log.Info("agent started", zap.String("server endpoint", cfg.EndpointHost),
		zap.Duration("pollInterval", cfg.PollInterval), zap.Duration("reportInterval", cfg.ReportInterval))

	// context for receiving os signals
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

	// save metrics before shutdown
	err = client.SendReport()
	if err != nil {
		log.Info("failed to save metrics to the server before shutdown")
	}
	log.Info("metrics saved successfully")

	log.Info("shutdown successfully")
}

// runReportSend starts the job of sending the collected metrics to the server.
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
