package conf

import (
	"trellis.tech/trellis/common.v1/config"
	"trellis.tech/trellis/common.v1/types"
)

type Config struct {
	Exporter ExporterConfig `yaml:"exporter" json:"exporter"`

	Inputs []*InputsConfig `yaml:"inputs" json:"inputs"`
	Output *OutputConfig   `yaml:"output" json:"output"`
}

type ExporterConfig struct {
	// 0 command; 1 server
	CommandType int               `yaml:"command_type" json:"command_type"`
	GlobalTags  map[string]string `yaml:"global_tags" json:"global_tags"`

	FlushInterval     types.Duration `yaml:"flush_interval" json:"flush_interval"`
	MetricBufferLimit int64          `yaml:"metric_buffer_limit" json:"metric_buffer_limit"`
	MetricBatchSize   int64          `yaml:"metric_batch_size" json:"metric_batch_size"`
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

func (p *Config) check() error {
	return nil
}

func GetConfigWithFile(filename string) (*Config, error) {
	c, err := config.NewConfig(filename)
	if err != nil {
		return nil, err
	}
	ec := &Config{}
	if err = c.ToObject("", ec); err != nil {
		return nil, err
	}

	if err = ec.check(); err != nil {
		return nil, err
	}
	return ec, nil
}
