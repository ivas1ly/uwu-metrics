package agent

import (
	"context"
	"net/url"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"go.uber.org/zap"

	"github.com/ivas1ly/uwu-metrics/internal/lib/logger"
	"github.com/ivas1ly/uwu-metrics/internal/utils"
)

const (
	reportMapSize  = 28
	minRandomValue = 100
	maxRandomValue = 100000
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

	metrics := &Metrics{}

	client := &Client{
		URL:     endpoint.String(),
		Metrics: metrics,
		Logger:  log,
		Key:     []byte(cfg.Key),
	}
	log.Info("agent started", zap.String("server endpoint", cfg.EndpointHost),
		zap.Duration("pollInterval", cfg.PollInterval), zap.Duration("reportInterval", cfg.ReportInterval))

	notifyCtx, stop := signal.NotifyContext(withCancel, os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)
	defer stop()

	go runMetricsUpdate(notifyCtx, metrics, cfg.PollInterval, log)

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

func runMetricsUpdate(ctx context.Context, metrics *Metrics, pollInterval time.Duration, log *zap.Logger) {
	updateTicker := time.NewTicker(pollInterval)
	defer updateTicker.Stop()

	log.Info("start update metrics job", zap.Duration("interval", pollInterval))

	for {
		select {
		case <-ctx.Done():
			log.Info("received done context")
			return
		case ut := <-updateTicker.C:
			metrics.UpdateMetrics()
			log.Info("[update] metrics updated", zap.Time("updated at", ut))
		}
	}
}

type Metrics struct {
	// Gauge
	MemStats    runtime.MemStats
	RandomValue float64
	// Counter
	PollCount int64
}

func (ms *Metrics) UpdateMetrics() {
	runtime.ReadMemStats(&ms.MemStats)

	ms.RandomValue = utils.RandFloat(minRandomValue, maxRandomValue)
	ms.PollCount++

	zap.L().Info("all metrics updated")
}

func (ms *Metrics) PrepareGaugeReport() map[string]float64 {
	report := make(map[string]float64, reportMapSize)

	report["Alloc"] = float64(ms.MemStats.Alloc)
	report["BuckHashSys"] = float64(ms.MemStats.BuckHashSys)
	report["Frees"] = float64(ms.MemStats.Frees)
	report["GCCPUFraction"] = ms.MemStats.GCCPUFraction
	report["GCSys"] = float64(ms.MemStats.GCSys)
	report["HeapAlloc"] = float64(ms.MemStats.HeapAlloc)
	report["HeapIdle"] = float64(ms.MemStats.HeapIdle)
	report["HeapInuse"] = float64(ms.MemStats.HeapInuse)
	report["HeapObjects"] = float64(ms.MemStats.HeapObjects)
	report["HeapReleased"] = float64(ms.MemStats.HeapReleased)
	report["HeapSys"] = float64(ms.MemStats.HeapSys)
	report["LastGC"] = float64(ms.MemStats.LastGC)
	report["Lookups"] = float64(ms.MemStats.Lookups)
	report["MCacheInuse"] = float64(ms.MemStats.MCacheInuse)
	report["MCacheSys"] = float64(ms.MemStats.MCacheSys)
	report["MSpanInuse"] = float64(ms.MemStats.MSpanInuse)
	report["MSpanSys"] = float64(ms.MemStats.MSpanSys)
	report["Mallocs"] = float64(ms.MemStats.Mallocs)
	report["NextGC"] = float64(ms.MemStats.NextGC)
	report["NumForcedGC"] = float64(ms.MemStats.NumForcedGC)
	report["NumGC"] = float64(ms.MemStats.NumGC)
	report["OtherSys"] = float64(ms.MemStats.OtherSys)
	report["PauseTotalNs"] = float64(ms.MemStats.PauseTotalNs)
	report["StackInuse"] = float64(ms.MemStats.StackInuse)
	report["StackSys"] = float64(ms.MemStats.StackSys)
	report["Sys"] = float64(ms.MemStats.Sys)
	report["TotalAlloc"] = float64(ms.MemStats.TotalAlloc)

	report["RandomValue"] = ms.RandomValue

	return report
}

func (ms *Metrics) PrepareCounterReport() map[string]int64 {
	report := make(map[string]int64, 1)
	report["PollCount"] = ms.PollCount

	return report
}
