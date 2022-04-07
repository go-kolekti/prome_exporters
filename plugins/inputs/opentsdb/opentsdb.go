package opentsdb

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"time"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	dto "github.com/prometheus/client_model/go"
	"trellis.tech/kolekti/prome_exporters/internal"
	"trellis.tech/kolekti/prome_exporters/plugins"
	"trellis.tech/kolekti/prome_exporters/plugins/inputs"
	"trellis.tech/trellis/common.v1/crypto/tls"
	"trellis.tech/trellis/common.v1/json"
	"trellis.tech/trellis/common.v1/types"
)

const (
	maxErrMsgLen   = 1024
	defaultTimeout = 10 * time.Second
)

var (
	instanceLabel = "instance"

	reg = regexp.MustCompile("[.-]")
)

const sampleConfig = `
`

type Collector struct {
	client *http.Client `yaml:"-" json:"-"`

	logger log.Logger `yaml:"-" json:"-"`

	Timeout types.Duration `yaml:"timeout" json:"timeout"`
	// ServerTypeConfig
	Schema      string `yaml:"schema" json:"schema"`
	Host        string `yaml:"host" json:"host"`
	Port        string `yaml:"port" json:"port"`
	MetricsPath string `yaml:"metrics_path" json:"metrics_path"`

	TlsConfig *tls.Config `yaml:"tls_config" json:"tls_config"`

	IgnoreMetricTimestamp bool `yaml:"ignore_metric_timestamp" json:"ignore_metric_timestamp"`
}

// SampleConfig returns the sample config
func (*Collector) SampleConfig() string {
	return sampleConfig
}

// Description returns the description
func (*Collector) Description() string {
	return ``
}

func timestampMsFunc(t int64) int64 {
	switch len(strconv.Itoa(int(t))) {
	case 10:
		return t * 1000
	case 13:
		return t
	case 16:
		return t / 1000
	default:
		return time.Now().Unix() / 1e6
	}
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

	ms := metrics{}

	if err := json.Unmarshal(bs, &ms); err != nil {
		return nil, err
	}

	metricFamilies := make(map[string]*dto.MetricFamily)

	for _, m := range ms {
		metricName := reg.ReplaceAllString(m.Metric, "_")
		mf, ok := metricFamilies[metricName]
		if !ok {
			typ := dto.MetricType_UNTYPED
			mf = &dto.MetricFamily{
				Name: &metricName,
				Type: &typ,
			}
		}
		th, _ := m.Value.Float64()

		metric := &dto.Metric{
			Untyped: &dto.Untyped{
				Value: &th,
			},
		}
		if !p.IgnoreMetricTimestamp {
			timestampMs := timestampMsFunc(m.Timestamp)
			metric.TimestampMs = &timestampMs
		}

		for k, v := range m.Tags {
			key, value := k, v
			metric.Label = append(metric.Label, &dto.LabelPair{Name: &key, Value: &value})
		}
		metric.Label = append(metric.Label, &dto.LabelPair{Name: &instanceLabel, Value: &urlP.Host})
		mf.Metric = append(mf.Metric, metric)

		metricFamilies[metricName] = mf
	}

	var metrics []*dto.MetricFamily
	for _, mf := range metricFamilies {
		metrics = append(metrics, mf)
	}
	return metrics, nil
}

type metrics []struct {
	Metric    string            `json:"metric"`
	Timestamp int64             `json:"timestamp"`
	Value     json.Number       `json:"value"`
	Tags      map[string]string `json:"tags"`
}

func init() {
	inputs.RegisterFactory("opentsdb", func(opts ...plugins.Option) (plugins.InputMetricsCollector, error) {

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
			p.Port = "4242"
		}

		if p.MetricsPath == "" {
			p.MetricsPath = "/api/stats"
		}

		return p, nil
	})
}
