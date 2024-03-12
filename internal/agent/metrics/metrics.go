package metrics

import (
	"context"
	"fmt"
	"runtime"
	"time"

	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/mem"
	"go.uber.org/zap"

	"github.com/ivas1ly/uwu-metrics/internal/utils/randfloat"
)

const (
	GaugeType      = "gauge"
	CounterType    = "counter"
	reportMapSize  = 50
	minRandomValue = 100
	maxRandomValue = 100000
)

// Metrics structure for storing the values of the collected metrics.
type Metrics struct {
	// Gauge
	UtilizationPerCPU []float64
	MemStats          runtime.MemStats
	RandomValue       float64
	TotalMemory       float64
	FreeMemory        float64
	// Counter
	PollCount int64
}

// UpdateMetrics gets the current metrics values and updates them in the Metrics structure.
func (ms *Metrics) UpdateMetrics() {
	runtime.ReadMemStats(&ms.MemStats)

	ms.RandomValue = randfloat.RandFloat(minRandomValue, maxRandomValue)
	ms.PollCount++

	zap.L().Info("all metrics updated")
}

// UpdatePsutilMetrics gets CPU and memory metrics values from the github.com/shirou/gopsutil package.
func (ms *Metrics) UpdatePsutilMetrics() error {
	v, err := mem.VirtualMemory()
	if err != nil {
		return err
	}

	ms.TotalMemory = float64(v.Total)
	ms.FreeMemory = float64(v.Free)

	ms.UtilizationPerCPU, err = cpu.Percent(0, true)
	if err != nil {
		return err
	}

	zap.L().Info("psutil metrics updated")

	return nil
}

// PrepareGaugeReport prepares the gauge type metrics to be sent to the server.
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

	report["TotalMemory"] = ms.TotalMemory
	report["FreeMemory"] = ms.FreeMemory

	for cpuNum, utilization := range ms.UtilizationPerCPU {
		report[fmt.Sprintf("CPUutilization%d", cpuNum)] = utilization
	}

	return report
}

// PrepareCounterReport prepares the counter type metrics to be sent to the server.
func (ms *Metrics) PrepareCounterReport() map[string]int64 {
	report := make(map[string]int64, 1)
	report["PollCount"] = ms.PollCount

	return report
}

// RunMetricsUpdate runs periodic updates to the metrics values.
func RunMetricsUpdate(ctx context.Context, metrics *Metrics, pollInterval time.Duration, log *zap.Logger) {
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

// RunPsutilMetricsUpdate runs periodic updates of metrics values from the github.com/shirou/gopsutil package.
func RunPsutilMetricsUpdate(ctx context.Context, metrics *Metrics, pollInterval time.Duration, log *zap.Logger) {
	updateTicker := time.NewTicker(pollInterval)
	defer updateTicker.Stop()

	log.Info("start update metrics job", zap.Duration("interval", pollInterval))

	for {
		select {
		case <-ctx.Done():
			log.Info("received done context")
			return
		case ut := <-updateTicker.C:
			err := metrics.UpdatePsutilMetrics()
			if err != nil {
				log.Info("[ERROR] can't get psutil metrics", zap.Error(err))
				continue
			}
			log.Info("[update] psutil metrics updated", zap.Time("updated at", ut))
		}
	}
}
