package agent

import (
	"log"
	"log/slog"
	"net/url"
	"os"
	"runtime"
	"time"

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

type Gauge float64
type Counter int64

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

func (ms *Metrics) PrepareGaugeReport() map[string]Gauge {
	report := make(map[string]Gauge, reportMapSize)

	report["Alloc"] = Gauge(ms.MemStats.Alloc)
	report["BuckHashSys"] = Gauge(ms.MemStats.BuckHashSys)
	report["Frees"] = Gauge(ms.MemStats.Frees)
	report["GCCPUFraction"] = Gauge(ms.MemStats.GCCPUFraction)
	report["GCSys"] = Gauge(ms.MemStats.GCSys)
	report["HeapAlloc"] = Gauge(ms.MemStats.HeapAlloc)
	report["HeapIdle"] = Gauge(ms.MemStats.HeapIdle)
	report["HeapInuse"] = Gauge(ms.MemStats.HeapInuse)
	report["HeapObjects"] = Gauge(ms.MemStats.HeapObjects)
	report["HeapReleased"] = Gauge(ms.MemStats.HeapReleased)
	report["HeapSys"] = Gauge(ms.MemStats.HeapSys)
	report["LastGC"] = Gauge(ms.MemStats.LastGC)
	report["Lookups"] = Gauge(ms.MemStats.Lookups)
	report["MCacheInuse"] = Gauge(ms.MemStats.MCacheInuse)
	report["MCacheSys"] = Gauge(ms.MemStats.MCacheSys)
	report["MSpanInuse"] = Gauge(ms.MemStats.MSpanInuse)
	report["MSpanSys"] = Gauge(ms.MemStats.MSpanSys)
	report["Mallocs"] = Gauge(ms.MemStats.Mallocs)
	report["NextGC"] = Gauge(ms.MemStats.NextGC)
	report["NumForcedGC"] = Gauge(ms.MemStats.NumForcedGC)
	report["NumGC"] = Gauge(ms.MemStats.NumGC)
	report["OtherSys"] = Gauge(ms.MemStats.OtherSys)
	report["PauseTotalNs"] = Gauge(ms.MemStats.PauseTotalNs)
	report["StackInuse"] = Gauge(ms.MemStats.StackInuse)
	report["StackSys"] = Gauge(ms.MemStats.StackSys)
	report["Sys"] = Gauge(ms.MemStats.Sys)
	report["TotalAlloc"] = Gauge(ms.MemStats.TotalAlloc)

	report["RandomValue"] = Gauge(ms.RandomValue)

	return report
}

func (ms *Metrics) PrepareCounterReport() map[string]Counter {
	report := make(map[string]Counter, 1)
	report["PollCount"] = Counter(ms.PollCount)

	return report
}
