package prometheus

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"mime"
	"net"
	"net/http"
	"net/url"
	"time"

	"trellis.tech/kolekti/prome_exporters/internal"
	"trellis.tech/kolekti/prome_exporters/plugins"
	"trellis.tech/kolekti/prome_exporters/plugins/inputs"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/matttproud/golang_protobuf_extensions/pbutil"
	dto "github.com/prometheus/client_model/go"
	"github.com/prometheus/common/expfmt"
	"trellis.tech/trellis/common.v1/crypto/tls"
	"trellis.tech/trellis/common.v1/types"
)

const (
	maxErrMsgLen   = 1024
	defaultTimeout = 10 * time.Second
)

var (
	instanceLabel = "instance"
)

const sampleConfig = `
name: prometheus
interval: 5s
options:
  host: 127.0.0.1
  port: 9090
  metrics_path: /metrics
  tags:
	haha: test
`

type Collector struct {
	client *http.Client
	logger log.Logger

	Timeout types.Duration `yaml:"timeout" json:"timeout"`
	// ServerTypeConfig
	Schema      string `yaml:"schema" json:"schema"`
	Host        string `yaml:"host" json:"host"`
	Port        string `yaml:"port" json:"port"`
	MetricsPath string `yaml:"metrics_path" json:"metrics_path"`

	TlsConfig *tls.Config `yaml:"tls_config" json:"tls_config"`

	Tags map[string]string `yaml:"tags" json:"tags"`
}

// SampleConfig returns the sample config
func (*Collector) SampleConfig() string {
	return sampleConfig
}

// Description returns the description
func (*Collector) Description() string {
	return ``
}

// Gather ...
func (p *Collector) Gather() ([]*dto.MetricFamily, error) {

	urlP := &url.URL{
		Scheme: p.Schema,
		Host:   net.JoinHostPort(p.Host, p.Port),
		Path:   p.MetricsPath,
	}

	req, _ := http.NewRequest("GET", urlP.String(), nil)
	resp, err := p.client.Do(req)
	if err != nil {
		// todo log
		return nil, err
	}
	defer internal.IOClose(resp.Body)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		errorLine := ""
		scanner := bufio.NewScanner(io.LimitReader(resp.Body, maxErrMsgLen))
		if scanner.Scan() {
			errorLine = scanner.Text()
		}
		err := fmt.Errorf("when writing to [%s] received status code: %d. body: %s", urlP.String(), resp.StatusCode, errorLine)
		level.Error(p.logger).Log("error", err)
		return nil, err
	}

	bs, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	metricFamilies := make(map[string]*dto.MetricFamily)

	var parser expfmt.TextParser
	// Read raw data
	buffer := bytes.NewBuffer(bs)
	reader := bufio.NewReader(buffer)

	mediatype, params, err := mime.ParseMediaType(resp.Header.Get("Content-Type"))
	if err == nil && mediatype == "application/vnd.google.protobuf" &&
		params["encoding"] == "delimited" &&
		params["proto"] == "io.prometheus.client.MetricFamily" {
		for {
			mf := &dto.MetricFamily{}
			if _, ierr := pbutil.ReadDelimited(reader, mf); ierr != nil {
				if ierr == io.EOF {
					break
				}

				level.Error(p.logger).Log("error",
					fmt.Errorf("reading metric family protocol buffer failed: %s", ierr))
			}
			metricFamilies[mf.GetName()] = mf
		}
	} else {

		metricFamilies, err = parser.TextToMetricFamilies(reader)
		if err != nil {
			level.Error(p.logger).Log("error", fmt.Errorf("reading text format failed: %s", err))
		}
	}

	var metrics []*dto.MetricFamily
	for _, mf := range metricFamilies {
		for i, metric := range mf.GetMetric() {
			metric.Label = append(metric.Label, &dto.LabelPair{Name: &instanceLabel, Value: &urlP.Host})

			for k, v := range p.Tags {
				key, value := k, v //
				metric.Label = append(metric.Label, &dto.LabelPair{Name: &key, Value: &value})
			}

			mf.GetMetric()[i] = metric
		}
		metrics = append(metrics, mf)
	}

	return metrics, nil
}

func init() {
	inputs.RegisterFactory("prometheus", func(opts ...plugins.Option) (plugins.InputMetricsCollector, error) {

		options := &plugins.Options{}
		for _, o := range opts {
			o(options)
		}

		p := &Collector{
			logger: options.Logger,
		}

		if options.Config != nil {
			if err := options.Config.ToObject("", p); err != nil {
				return nil, err
			}
		}

		timeout := defaultTimeout
		if p.Timeout != 0 {
			timeout = time.Duration(p.Timeout)
		}

		transport := &http.Transport{
			Proxy: http.ProxyFromEnvironment,
		}

		if p.TlsConfig != nil {
			tlsConfig, err := p.TlsConfig.GetTLSConfig()
			if err != nil {
				return nil, err
			}
			transport.TLSClientConfig = tlsConfig
		}

		p.client = &http.Client{
			Timeout:   timeout,
			Transport: transport,
		}

		if p.Schema == "" {
			p.Schema = "http"
		}

		if p.Host == "" {
			p.Host = "127.0.0.1"
		}

		if p.Port == "" {
			p.Port = "80"
		}

		if p.MetricsPath == "" {
			p.MetricsPath = "/metrics"
		}

		return p, nil
	})
}
