package metrics

import "github.com/prometheus/client_golang/prometheus"

// MillisecondBuckets covers 1ms to ~32s for histograms observed in milliseconds;
// prometheus.DefBuckets assumes seconds and caps at 10, saturating ms values.
var MillisecondBuckets = prometheus.ExponentialBuckets(1, 2, 16)
