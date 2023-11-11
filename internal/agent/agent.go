package agent

import (
	"log"
	"log/slog"
	"net/url"
	"os"
	"runtime"
	"time"

	"github.com/ivas1ly/uwu-metrics/internal/agent/entity"
	"github.com/ivas1ly/uwu-metrics/internal/utils"
)

const (
	reportMapSize  = 28
	minRandomValue = 100
	maxRandomValue = 100000
)

func Run(cfg *Config) {
	opts := &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}
	logger := slog.New(slog.NewTextHandler(os.Stdout, opts)).
		With(slog.String("app", "agent"))
	slog.SetDefault(logger)

	metricsUpdateTicker := time.NewTicker(cfg.PollInterval)
	reportSendTicker := time.NewTicker(cfg.ReportInterval)

	defer metricsUpdateTicker.Stop()
	defer reportSendTicker.Stop()

	endpoint := url.URL{
		Scheme: "http",
		Host:   cfg.EndpointHost,
		Path:   "/update/",
	}

	metrics := &Metrics{}
	client := Client{
		URL:     endpoint.String(),
		Metrics: metrics,
		Logger:  logger,
	}
	logger.Info("agent started", slog.String("server endpoint", cfg.EndpointHost),
		slog.Duration("pollInterval", cfg.PollInterval), slog.Duration("reportInterval", cfg.ReportInterval))

	for {
		select {
		case mut := <-metricsUpdateTicker.C:
			logger.Info("[update] metrics updated", slog.Time("updated at", mut))
			metrics.UpdateMetrics()
		case rst := <-reportSendTicker.C:
			logger.Info("[report] metrics sent to server", slog.Time("sent at", rst))
			client.SendReport()
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

	log.Println("all metrics updated")
}

func (ms *Metrics) PrepareGaugeReport() map[string]entity.Gauge {
	report := make(map[string]entity.Gauge, reportMapSize)

	report["Alloc"] = entity.Gauge(ms.MemStats.Alloc)
	report["BuckHashSys"] = entity.Gauge(ms.MemStats.BuckHashSys)
	report["Frees"] = entity.Gauge(ms.MemStats.Frees)
	report["GCCPUFraction"] = entity.Gauge(ms.MemStats.GCCPUFraction)
	report["GCSys"] = entity.Gauge(ms.MemStats.GCSys)
	report["HeapAlloc"] = entity.Gauge(ms.MemStats.HeapAlloc)
	report["HeapIdle"] = entity.Gauge(ms.MemStats.HeapIdle)
	report["HeapInuse"] = entity.Gauge(ms.MemStats.HeapInuse)
	report["HeapObjects"] = entity.Gauge(ms.MemStats.HeapObjects)
	report["HeapReleased"] = entity.Gauge(ms.MemStats.HeapReleased)
	report["HeapSys"] = entity.Gauge(ms.MemStats.HeapSys)
	report["LastGC"] = entity.Gauge(ms.MemStats.LastGC)
	report["Lookups"] = entity.Gauge(ms.MemStats.Lookups)
	report["MCacheInuse"] = entity.Gauge(ms.MemStats.MCacheInuse)
	report["MCacheSys"] = entity.Gauge(ms.MemStats.MCacheSys)
	report["MSpanInuse"] = entity.Gauge(ms.MemStats.MSpanInuse)
	report["MSpanSys"] = entity.Gauge(ms.MemStats.MSpanSys)
	report["Mallocs"] = entity.Gauge(ms.MemStats.Mallocs)
	report["NextGC"] = entity.Gauge(ms.MemStats.NextGC)
	report["NumForcedGC"] = entity.Gauge(ms.MemStats.NumForcedGC)
	report["NumGC"] = entity.Gauge(ms.MemStats.NumGC)
	report["OtherSys"] = entity.Gauge(ms.MemStats.OtherSys)
	report["PauseTotalNs"] = entity.Gauge(ms.MemStats.PauseTotalNs)
	report["StackInuse"] = entity.Gauge(ms.MemStats.StackInuse)
	report["StackSys"] = entity.Gauge(ms.MemStats.StackSys)
	report["Sys"] = entity.Gauge(ms.MemStats.Sys)
	report["TotalAlloc"] = entity.Gauge(ms.MemStats.TotalAlloc)

	report["RandomValue"] = entity.Gauge(ms.RandomValue)

	return report
}

func (ms *Metrics) PrepareCounterReport() map[string]entity.Counter {
	report := make(map[string]entity.Counter, 1)
	report["PollCount"] = entity.Counter(ms.PollCount)

	return report
}
