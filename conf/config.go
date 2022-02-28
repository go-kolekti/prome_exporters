package conf

import (
	"trellis.tech/trellis/common.v1/config"
	"trellis.tech/trellis/common.v1/types"
)

type ExporterConfig struct {
	// 0 command; 1 server
	CommandType int `yaml:"command_type" json:"command_type"`

	FlushInterval types.Duration `yaml:"flush_interval" json:"flush_interval"`

	MetricBufferLimit int64 `yaml:"metric_buffer_limit" json:"metric_buffer_limit"`
	MetricBatchSize   int64 `yaml:"metric_batch_size" json:"metric_batch_size"`

	Inputs []*InputsConfig `yaml:"inputs" json:"inputs"`
	Output *OutputConfig   `yaml:"output" json:"output"`

	Tags map[string]string `yaml:"tags" json:"tags"`
}

type InputsConfig struct {
	Name     string         `yaml:"name" json:"name"`
	Interval types.Duration `yaml:"interval" json:"interval"`

	Tags map[string]string `yaml:"tags" json:"tags"`

	Options config.Options `json:"options" yaml:"options"`
}

type OutputConfig struct {
	Name    string         `yaml:"name" json:"name"`
	Options config.Options `json:"options" yaml:"options"`
}

func (p *ExporterConfig) check() error {
	return nil
}

func GetConfigWithFile(filename string) (*ExporterConfig, error) {
	c, err := config.NewConfig(filename)
	if err != nil {
		return nil, err
	}
	ec := &ExporterConfig{}
	if err = c.ToObject("", ec); err != nil {
		return nil, err
	}

	if err = ec.check(); err != nil {
		return nil, err
	}
	return ec, nil
}
