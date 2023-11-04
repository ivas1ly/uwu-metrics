package metrics

type Metric struct {
	Type  string
	Name  string
	Value string
}

type Metrics struct {
	Counter map[string]int64
	Gauge   map[string]float64
}
