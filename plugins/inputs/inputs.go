package inputs

import (
	"strings"

	dto "github.com/prometheus/client_model/go"

	"github.com/go-kit/log"
	"github.com/prometheus/client_golang/prometheus"

	"trellis.tech/kolekti/prome_exporters/plugins"
	"trellis.tech/trellis/common.v1/config"
	"trellis.tech/trellis/common.v1/errcode"
)

type InputType int32

const (
	InputTypePrometheusCollector InputType = iota
	InputTypeMetricsCollector
)

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

type FactoryPrometheusCollector func(...Option) (prometheus.Collector, error)
type FactoryMetricsCollector func(...Option) (InputMetricsCollector, error)

var (
	mapNewCollectorFunc = make(map[string]*Input)
)

type Input struct {
	inputType InputType

	pcFactory FactoryPrometheusCollector
	mcFactory FactoryMetricsCollector
}

func (p *Input) InputType() InputType {
	return p.inputType
}

func (p *Input) NewPrometheusCollector(opts ...Option) (InputPrometheusCollector, error) {
	if p.inputType != InputTypePrometheusCollector || p.pcFactory == nil {
		return nil, errcode.Newf("its not prometheus collector factory, type: %d, %+v", p.inputType, p.pcFactory)
	}

	return p.pcFactory(opts...)
}

func (p *Input) NewMetricsCollector(opts ...Option) (InputMetricsCollector, error) {
	if p.inputType != InputTypeMetricsCollector || p.mcFactory == nil {
		return nil, errcode.Newf("its not metrics collector factory, type: %d, %+v", p.inputType, p.mcFactory)
	}

	return p.mcFactory(opts...)
}

type InputPrometheusCollector interface {
	prometheus.Collector
}

type InputMetricsCollector interface {
	plugins.PluginDescriber
	Gather() ([]*dto.MetricFamily, error)
}

func RegisterFactory(name string, fn interface{}) {
	if name = strings.TrimSpace(name); name == "" {
		panic(errcode.New("empty collector name"))
	}
	if fn == nil {
		panic(errcode.New("nil collector factory"))
	}
	if _, ok := mapNewCollectorFunc[name]; ok {
		panic(errcode.Newf("collector factory already exists: %s", name))
	}

	var input = &Input{}
	switch t := fn.(type) {
	case FactoryPrometheusCollector:
		input.inputType = InputTypePrometheusCollector
		input.pcFactory = t
	case func(...Option) (prometheus.Collector, error):
		input.inputType = InputTypePrometheusCollector
		input.pcFactory = t
	case FactoryMetricsCollector:
		input.inputType = InputTypeMetricsCollector
		input.mcFactory = t
	case func(...Option) (InputMetricsCollector, error):
		input.inputType = InputTypeMetricsCollector
		input.mcFactory = t
	default:
		panic(errcode.Newf("not supported type : %+v", t))
	}
	mapNewCollectorFunc[name] = input
}

func GetFactory(name string) (*Input, error) {
	input, ok := mapNewCollectorFunc[name]
	if !ok || input == nil {
		return nil, errcode.Newf("not found collector factory: %s", name)
	}
	return input, nil
}
