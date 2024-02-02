package exporter

import (
	"context"

	"github.com/prometheus/client_golang/prometheus"
)

type Exporter interface {
	prometheus.Collector

	Init(ctx context.Context, subsystem string) error
	SubCollector() []prometheus.Collector
}

type Creator func() Exporter

var Collectors = make(map[string]Creator)

func AddCollectors(name string, col Creator) {
	Collectors[name] = col
}
