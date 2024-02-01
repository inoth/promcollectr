package nginx

import "github.com/prometheus/client_golang/prometheus"

type metrics struct {
	size     prometheus.Counter
	duration *prometheus.HistogramVec
	requests *prometheus.CounterVec
}
