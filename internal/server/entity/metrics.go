package entity

// Metrics is a structure for temporarily working with metrics.
type Metrics struct {
	Counter map[string]int64
	Gauge   map[string]float64
}

const (
	CounterType = "counter"
	GaugeType   = "gauge"
)
