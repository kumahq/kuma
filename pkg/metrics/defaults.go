package metrics

// DefaultObjectives defines default percentiles for Summary (50th, 90th, 99th percentile)
var DefaultObjectives = map[float64]float64{
	0.5:  0.05,
	0.9:  0.01,
	0.99: 0.001,
}
