package promcollectr

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/inoth/promcollectr/exporter"
	"github.com/inoth/toybox"
	"github.com/inoth/toybox/util"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/client_golang/prometheus/push"
)

type conf struct {
	Exporter map[string][]toml.Primitive `toml:"exporter"`
}

type PromcollectrServer struct {
	ready     bool
	name      string
	exporters []exporter.Exporter
	cfg       conf
	mate      toml.MetaData

	ServerName string `toml:"server_name"`
	Port       string `toml:"port"`
	Path       string `toml:"path"`
	CfgPath    string `toml:"cfg_path"`

	PushHost     string `toml:"push_host"`
	PushJob      string `toml:"push_job"`
	PushInterval int    `toml:"push_interval"`
}

func NewPromcollectrComponent(opts ...Option) toybox.Option {
	o := defaultOption()
	for _, opt := range opts {
		opt(&o)
	}
	return func(tb *toybox.ToyBox) {
		tb.AppendServer(&o)
	}
}

func (pm *PromcollectrServer) IsReady() {
	pm.ready = true
}

func (pm *PromcollectrServer) Ready() bool {
	return pm.ready
}

func (pm *PromcollectrServer) Name() string {
	return pm.name
}

func (pm *PromcollectrServer) push(ctx context.Context, reg prometheus.Gatherer) {
	tk := time.NewTicker(time.Second * time.Duration(pm.PushInterval))
	ph := push.New(pm.PushHost, pm.PushJob).Gatherer(reg)
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-tk.C:
				err := ph.Push()
				if err != nil {
					fmt.Printf("push failed: %s\n", err)
					continue
				}
			}
		}
	}()
}

func (pm *PromcollectrServer) Run(ctx context.Context) error {
	if !pm.ready {
		return fmt.Errorf("server %s not ready", pm.name)
	}

	if err := pm.loadExporterCfg(); err != nil {
		return errors.Wrap(err, "pm.loadExporterCfg() failed")
	}

	if err := pm.initExporter(ctx); err != nil {
		return errors.Wrap(err, "pm.initExporter() failed")
	}

	reg, err := pm.register()
	if err != nil {
		return err
	}

	pm.push(ctx, reg)

	mux := http.NewServeMux()
	mux.Handle(pm.Path, promhttp.HandlerFor(reg, promhttp.HandlerOpts{Registry: reg}))

	if err := http.ListenAndServe(pm.Port, mux); err != nil {
		return errors.Wrap(err, "http.ListenAndServe failed")
	}
	return nil
}

func (pm *PromcollectrServer) loadExporterCfg() error {
	paths, err := util.PathGlobPattern(pm.CfgPath + "/*.toml")
	if err != nil {
		panic(fmt.Errorf("no configuration available"))
	}
	var sb strings.Builder
	for _, path := range paths {
		buf, err := util.ReadFile(path)
		if err != nil {
			fmt.Printf("%s read file err: %v", path, err)
			continue
		}
		sb.Write(buf)
		sb.WriteString("\n")
	}

	tomlStr := sb.String()
	pm.mate, err = toml.Decode(tomlStr, &(pm.cfg))
	if err != nil {
		panic(fmt.Errorf("toml.Decode: %v", err))
	}
	if err := pm.loadExporter(); err != nil {
		panic(err)
	}
	return nil
}

func (pm *PromcollectrServer) loadExporter() error {
	for key, val := range pm.cfg.Exporter {
		if greator, ok := exporter.Collectors[key]; ok {
			for _, v := range val {
				col := greator()
				if err := pm.mate.PrimitiveDecode(v, col); err != nil {
					return errors.Wrap(err, "pm.mate.PrimitiveDecode filed: "+key)
				}
				pm.exporters = append(pm.exporters, col)
			}
		}
	}
	return nil
}

func (pm *PromcollectrServer) initExporter(ctx context.Context) error {
	for _, exp := range pm.exporters {
		if err := exp.Init(ctx, pm.ServerName); err != nil {
			return err
		}
	}
	return nil
}

func (pm *PromcollectrServer) register() (*prometheus.Registry, error) {
	reg := prometheus.NewRegistry()
	for _, item := range pm.exporters {
		if val, ok := item.(prometheus.Collector); ok {
			reg.MustRegister(val)
			reg.MustRegister(item.SubCollector()...)
		}
	}
	return reg, nil
}
