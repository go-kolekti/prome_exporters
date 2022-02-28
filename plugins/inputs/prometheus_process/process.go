package prometheus_process

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"trellis.tech/kolekti/prome_exporters/plugins/inputs"
)

func init() {
	inputs.RegisterFactory("prometheus_process", NewProcessCollector)
}

// Config defines the behavior of a process metrics collector
// created with Config.
type Config struct {
	// If non-empty, each of the collected metrics is prefixed by the
	// provided string and an underscore ("_").
	Namespace string `yaml:"namespace"`
	// If true, any error encountered during collection is reported as an
	// invalid metric (see NewInvalidMetric). Otherwise, errors are ignored
	// and the collected metrics will be incomplete. (Possibly, no metrics
	// will be collected at all.) While that's usually not desired, it is
	// appropriate for the common "mix-in" of process metrics, where process
	// metrics are nice to have, but failing to collect them should not
	// disrupt the collection of the remaining metrics.
	ReportErrors bool `yaml:"report_errors"`
}

func NewProcessCollector(opts ...inputs.Option) (prometheus.Collector, error) {
	options := &inputs.Options{}
	for _, opt := range opts {
		opt(options)
	}
	c := &Config{}
	if options.Config != nil {
		if err := options.Config.ToObject("", c); err != nil {
			return nil, err
		}
	}
	return collectors.NewProcessCollector(collectors.ProcessCollectorOpts{
		Namespace:    c.Namespace,
		ReportErrors: c.ReportErrors,
	}), nil
}
