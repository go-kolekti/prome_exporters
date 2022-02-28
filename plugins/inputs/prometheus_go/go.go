package prometheus_go

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"trellis.tech/kolekti/prome_exporters/plugins/inputs"
)

func init() {
	inputs.RegisterFactory("prometheus_go", NewGoCollector)
}
func NewGoCollector(...inputs.Option) (prometheus.Collector, error) {
	return collectors.NewGoCollector(), nil
}
