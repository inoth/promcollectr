package nginx

import (
	"fmt"

	"github.com/inoth/promcollectr/exporter"
	"github.com/prometheus/client_golang/prometheus"
)

type NginxCollector struct {
	Name    string `toml:"name"`
	Addr    string `toml:"addr"`
	LogPath string `toml:"log_path"`
}

func (nc *NginxCollector) Init() error {
	fmt.Println("nginx hello world: " + nc.Name)
	return nil
}

func (nc *NginxCollector) Describe(chan<- *prometheus.Desc) {

}

func (nc *NginxCollector) Collect(chan<- prometheus.Metric) {

}

func init() {
	exporter.AddCollectors("nginx", func() exporter.Exporter {
		return &NginxCollector{}
	})
}
