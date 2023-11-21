package entity

type Metrics struct {
	Counter map[string]int64
	Gauge   map[string]float64
}

const (
	CounterType = "counter"
	GaugeType   = "gauge"
)
