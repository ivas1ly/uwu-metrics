package agent

import (
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
	log := logger.New(defaultLogLevel).
		With(zap.String("app", "agent"))

	metricsUpdateTicker := time.NewTicker(cfg.PollInterval)
	reportSendTicker := time.NewTicker(cfg.ReportInterval)

	endpoint := url.URL{
		Scheme: "http",
		Host:   cfg.EndpointHost,
		Path:   "/updates/",
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
				err := client.SendReport()
				if err != nil {
					log.Info("[report] failed to send metrics to server")
				}
			case <-done:
				log.Info("all tickers have been stopped")
				return
			}
		}
	}()

	// block until signal is received
	sig := <-c
	log.Info("app got os signal", zap.String("signal", sig.String()))
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
