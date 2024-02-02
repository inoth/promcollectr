package exporter

import "github.com/prometheus/client_golang/prometheus"

type Exporter interface {
	prometheus.Collector

	// Name() string
	Init() error
}

type Creator func() Exporter

var Collectors = make(map[string]Creator)

func AddCollectors(name string, col Creator) {
	Collectors[name] = col
}
