package nginx

import (
	"context"
	"fmt"
	"log"
	"strconv"

	"github.com/hpcloud/tail"
	"github.com/inoth/toybox/util"
	"github.com/prometheus/client_golang/prometheus"
)

type metrics struct {
	size     prometheus.Counter
	duration *prometheus.HistogramVec
	requests *prometheus.CounterVec
}

func NewMetrics(nc *NginxCollector, subsystem, namespace string) *metrics {
	m := &metrics{
		size: prometheus.NewCounter(prometheus.CounterOpts{
			Subsystem: subsystem,
			Namespace: namespace,
			Name:      "size_bytes_total",
			Help:      "Total bytes sent to the clients.",
		}),
		requests: prometheus.NewCounterVec(prometheus.CounterOpts{
			Subsystem: subsystem,
			Namespace: namespace,
			Name:      "http_requests_total",
			Help:      "Total number of requests.",
		}, []string{"status_code", "method", "uri"}),
		duration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Subsystem: subsystem,
			Namespace: namespace,
			Name:      "http_request_duration_seconds",
			Help:      "Duration of the request.",
			Buckets:   prometheus.DefBuckets,
		}, []string{"status_code", "method", "uri"}),
	}

	nc.collectors = append(nc.collectors, m.size, m.requests, m.duration)
	return m
}

func (m *metrics) tailAccessLogFile(ctx context.Context, path string) {
	t, err := tail.TailFile(path, tail.Config{Follow: true, ReOpen: true})
	if err != nil {
		log.Fatalf("tail.TailFile failed: %s", err)
	}
	for line := range t.Lines {
		select {
		case <-ctx.Done():
			return
		default:
			res, err := util.JsonParse[map[string]any]([]byte(line.Text))
			if err != nil {
				continue
			}

			result := convertFieldsToString(res)

			s, err := strconv.ParseFloat(result["bytes"], 64)
			if err != nil {
				continue
			}
			m.size.Add(s)

			m.requests.With(prometheus.Labels{"method": result["method"], "status_code": result["status"], "uri": result["uri"]}).Add(1)

			u, err := strconv.ParseFloat(result["request_time"], 64)
			if err != nil {
				continue
			}
			m.duration.With(prometheus.Labels{"method": result["method"], "status_code": result["status"], "uri": result["uri"]}).Observe(u)
		}
	}
}

func convertFieldsToString(data map[string]interface{}) map[string]string {
	result := make(map[string]string)
	for key, value := range data {
		switch v := value.(type) {
		case string:
			result[key] = v
		default:
			result[key] = fmt.Sprint(v)
		}
	}
	return result
}
