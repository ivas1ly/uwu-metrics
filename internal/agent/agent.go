package agent

import (
	"net/url"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"go.uber.org/zap"

	"github.com/ivas1ly/uwu-metrics/internal/agent/entity"
	"github.com/ivas1ly/uwu-metrics/internal/lib/logger"
	"github.com/ivas1ly/uwu-metrics/internal/utils"
)

const (
	reportMapSize  = 28
	minRandomValue = 100
	maxRandomValue = 100000
)

func Run(cfg *Config) {
	log := logger.New(defaultLogLevel).
		With(zap.String("app", "agent"))

	metricsUpdateTicker := time.NewTicker(cfg.PollInterval)
	reportSendTicker := time.NewTicker(cfg.ReportInterval)

	endpoint := url.URL{
		Scheme: "http",
		Host:   cfg.EndpointHost,
		Path:   "/update/",
	}

	metrics := &Metrics{}
	client := Client{
		URL:     endpoint.String(),
		Metrics: metrics,
		Logger:  log,
	}
	log.Info("agent started", zap.String("server endpoint", cfg.EndpointHost),
		zap.Duration("pollInterval", cfg.PollInterval), zap.Duration("reportInterval", cfg.ReportInterval))

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)

	done := make(chan bool)

	go func() {
		defer func() {
			done <- true
		}()
		for {
			select {
			case mut := <-metricsUpdateTicker.C:
				log.Info("[update] metrics updated", zap.Time("updated at", mut))
				metrics.UpdateMetrics()
			case rst := <-reportSendTicker.C:
				log.Info("[report] metrics sent to server", zap.Time("sent at", rst))
				client.SendReport()
			case <-done:
				log.Info("all tickers have been stopped")
				return
			}
		}
	}()

	// block until signal is received
	sig := <-c
	log.Warn("app got os signal", zap.String("signal", sig.String()))
	log.Info("gracefully shutting down...")
	reportSendTicker.Stop()
	metricsUpdateTicker.Stop()

	// stop goroutine
	done <- true

	// wait until done
	<-done

	log.Info("shutdown successfully")
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
