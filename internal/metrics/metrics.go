package metrics

type Metric struct {
	Counter int64
	Gauge   float64
	Type    string
}
