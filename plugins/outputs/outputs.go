package outputs

import (
	"github.com/go-kit/log"
	dto "github.com/prometheus/client_model/go"
	"trellis.tech/kolekti/prome_exporters/plugins"
	"trellis.tech/trellis/common.v1/config"
)

type Factory func(...Option) (Output, error)

type Option func(*Options)
type Options struct {
	Config config.Config
	Logger log.Logger
}

func Config(c config.Config) Option {
	return func(o *Options) {
		o.Config = c
	}
}

func Logger(l log.Logger) Option {
	return func(o *Options) {
		o.Logger = l
	}
}

type Output interface {
	plugins.PluginDescriber

	// Connect to the Output; connect is only called once when the plugin starts
	Connect() error
	// Close any connections to the Output. Close is called once when the output
	// is shutting down. Close will not be called until all writes have finished,
	// and Write() will not be called once Close() has been, so locking is not
	// necessary.
	Close() error
	// Write takes in group of points to be written to the Output
	Write(metrics []*dto.MetricFamily) error
}
