package promethues_node_exporter

import (
	"trellis.tech/kolekti/prome_exporters/internal"
	"trellis.tech/kolekti/prome_exporters/plugins"
	"trellis.tech/kolekti/prome_exporters/plugins/inputs"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/node_exporter/collector"
)

func init() {
	inputs.RegisterFactory("prometheus_node_exporter", NewNodeExporterCollector)
}

type Collector struct {
	prometheus.Collector
}

func (p *Collector) Tags() map[string]string {
	return map[string]string{
		"instance": internal.GetIP(),
	}
}

func NewNodeExporterCollector(opts ...plugins.Option) (_ plugins.InputPrometheusCollector, err error) {
	options := &plugins.Options{}
	for _, opt := range opts {
		opt(options)
	}

	var filters []string
	if options.Config != nil {
		filters = options.Config.GetStringList("filters")
	}

	c := &Collector{}
	c.Collector, err = collector.NewNodeCollector(options.Logger, filters...)
	if err != nil {
		return
	}

	return c, nil
}
