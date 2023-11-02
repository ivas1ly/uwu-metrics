package agent

import (
	"log"
	"log/slog"
	"math/rand"
	"os"
	"runtime"
	"time"
)

const (
	pollInterval   = 2
	reportInterval = 10
	reportURL      = "http://localhost:8080/update/"
)

func Run() {
	opts := &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}
	logger := slog.New(slog.NewTextHandler(os.Stdout, opts))
	slog.SetDefault(logger)

	metricsUpdateTicker := time.NewTicker(pollInterval * time.Second)
	reportSendTicker := time.NewTicker(reportInterval * time.Second)

	defer metricsUpdateTicker.Stop()
	defer reportSendTicker.Stop()

	metrics := Metrics{}
	client := Client{
		URL:     reportURL,
		Metrics: &metrics,
		Logger:  logger,
	}

	for {
		select {
		case mut := <-metricsUpdateTicker.C:
			logger.Info("[update] metrics updated at:", slog.Time("updated at", mut))
			metrics.UpdateMetrics()
		case rst := <-reportSendTicker.C:
			logger.Info("[report] metrics sent at:", slog.Time("report sent at", rst))
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

	ms.RandomValue = randFloat(1000, 1000000)
	ms.PollCount += 1

	log.Println("all metrics updated")
}

func (ms *Metrics) PrepareGaugeReport() map[string]Gauge {
	report := make(map[string]Gauge, 28)

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

func randFloat(min, max float64) float64 {
	return min + rand.Float64()*(max-min)
}
