package promcollectr

import (
	"github.com/inoth/promcollectr/exporter"
)

type PromcollectrComponent struct {
	exporters []exporter.Exporter
}
