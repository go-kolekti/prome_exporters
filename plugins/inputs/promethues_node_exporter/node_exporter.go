package promethues_node_exporter

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/node_exporter/collector"
	"trellis.tech/kolekti/prome_exporters/plugins/inputs"
)

func init() {
	inputs.RegisterFactory("prometheus_node_exporter", NewNodeExporterCollector)
}

func NewNodeExporterCollector(opts ...inputs.Option) (_ prometheus.Collector, err error) {
	options := &inputs.Options{}
	for _, opt := range opts {
		opt(options)
	}

	return collector.NewNodeCollector(options.Logger)
}
