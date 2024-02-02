package promcollectr

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/inoth/promcollectr/exporter"
	"github.com/inoth/toybox"
	"github.com/inoth/toybox/util"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"golang.org/x/sync/errgroup"
)

type conf struct {
	Exporter map[string][]toml.Primitive `toml:"exporter"`
}

type PromcollectrComponent struct {
	ready     bool
	name      string
	exporters []exporter.Exporter
	cfg       conf
	mate      toml.MetaData

	ServerName string `toml:"server_name"`
	Port       string `toml:"port"`
	Path       string `toml:"path"`
	CfgPath    string `toml:"cfg_path"`
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

func (pm *PromcollectrComponent) IsReady() {
	pm.ready = true
}

func (pm *PromcollectrComponent) Ready() bool {
	return pm.ready
}

func (pm *PromcollectrComponent) Name() string {
	return pm.name
}

func (pm *PromcollectrComponent) Run(ctx context.Context) error {
	if !pm.ready {
		return fmt.Errorf("server %s not ready", pm.name)
	}

	if err := pm.initExporter(); err != nil {
		return errors.Wrap(err, "pm.initExporter() failed")
	}

	pm.runExporter(ctx)

	reg, err := pm.register()
	if err != nil {
		return err
	}

	mux := http.NewServeMux()
	mux.Handle(pm.Path, promhttp.HandlerFor(reg, promhttp.HandlerOpts{Registry: reg}))
	if err := http.ListenAndServe(pm.Port, mux); err != nil {
		return errors.Wrap(err, "http.ListenAndServe failed")
	}

	return nil
}

func (pm *PromcollectrComponent) initExporter() error {
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

func (pm *PromcollectrComponent) loadExporter() error {
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

func (pm *PromcollectrComponent) runExporter(ctx context.Context) {
	eg, egctx := errgroup.WithContext(ctx)
	for _, exp := range pm.exporters {
		exp := exp
		eg.Go(func() error {
			return exp.Init(egctx, pm.ServerName)
		})
	}
	if err := eg.Wait(); err != nil {
		panic(errors.Wrap(err, "run exporter init failed"))
	}
}

func (pm *PromcollectrComponent) register() (*prometheus.Registry, error) {
	reg := prometheus.NewRegistry()
	for _, item := range pm.exporters {
		if val, ok := item.(prometheus.Collector); ok {
			reg.MustRegister(val)
		}
	}
	return reg, nil
}
