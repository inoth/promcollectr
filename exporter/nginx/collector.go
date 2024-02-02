package nginx

import (
	"bytes"
	"context"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/inoth/promcollectr/exporter"
	"github.com/prometheus/client_golang/prometheus"
)

type NginxCollector struct {
	Name    string `toml:"name"`
	Addr    string `toml:"addr"`
	LogPath string `toml:"log_path"`

	collectors []prometheus.Collector

	stats             func() ([]NginxStats, error)
	ConnectionsActive *prometheus.Desc `toml:"-"`
	Connections       *prometheus.Desc `toml:"-"`
}

func (nc *NginxCollector) Init(ctx context.Context, namespace string) error {
	nc.ConnectionsActive = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, nc.Name, "nginx_connections_active"),
		"Number of active connections.",
		[]string{},
		nil,
	)
	nc.Connections = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, nc.Name, "nginx_connections_total"),
		"Connections (Reading - Writing - Waiting)",
		[]string{"type"},
		nil,
	)
	nc.stats = func() ([]NginxStats, error) {
		var client = &http.Client{
			Timeout: time.Second * 10,
		}
		resp, err := client.Get(nc.Addr)
		if err != nil {
			log.Fatalf("client.Get failed %s: %s", nc.Addr, err)
		}
		defer resp.Body.Close()
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Fatalf("io.ReadAll failed %s", err)
		}
		r := bytes.NewReader(body)
		return ScanBasicStats(r)
	}

	metric := NewMetrics(namespace, nc.Name)

	nc.collectors = append(nc.collectors,
		metric.size,
		metric.requests,
		metric.duration,
	)

	go metric.tailAccessLogFile(ctx, nc.LogPath)

	return nil
}

func (nc *NginxCollector) SubCollector() []prometheus.Collector {
	return nc.collectors
}

func (nc *NginxCollector) Describe(ch chan<- *prometheus.Desc) {
	ds := []*prometheus.Desc{
		nc.ConnectionsActive,
	}
	for _, d := range ds {
		ch <- d
	}
}

func (nc *NginxCollector) Collect(ch chan<- prometheus.Metric) {
	stats, err := nc.stats()
	if err != nil {
		ch <- prometheus.NewInvalidMetric(nc.ConnectionsActive, err)
		return
	}
	for _, s := range stats {
		ch <- prometheus.MustNewConstMetric(
			nc.ConnectionsActive,
			prometheus.GaugeValue,
			s.ConnectionsActive,
		)
		for _, conn := range s.Connections {
			conns := []struct {
				connType string
				total    float64
			}{
				{connType: conn.Type, total: conn.Total},
			}
			for _, connT := range conns {
				ch <- prometheus.MustNewConstMetric(
					nc.Connections,
					prometheus.CounterValue,
					connT.total,
					connT.connType,
				)
			}
		}
	}
}

func init() {
	exporter.AddCollectors("nginx", func() exporter.Exporter {
		return &NginxCollector{}
	})
}
