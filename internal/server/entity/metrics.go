package entity

const (
	CounterType = "counter"
	GaugeType   = "gauge"
)

// Metrics is a structure for temporarily working with metrics.
type Metrics struct {
	Counter map[string]int64
	Gauge   map[string]float64
}

type Metric struct {
	Delta *int64
	Value *float64
	ID    string
	MType string
}
