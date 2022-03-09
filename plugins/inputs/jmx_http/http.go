package jmx_http

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"trellis.tech/kolekti/prome_exporters/plugins"
	"trellis.tech/kolekti/prome_exporters/plugins/inputs"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	dto "github.com/prometheus/client_model/go"
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
`

type HTTPMetricCollector struct {
	client *http.Client `yaml:"-" json:"-"`

	logger log.Logger `yaml:"-" json:"-"`

	Timeout types.Duration `toml:"timeout"`
	// ServerTypeConfig
	Schema      string `yaml:"schema" json:"schema"`
	Host        string `yaml:"host" json:"host"`
	Port        string `yaml:"port" json:"port"`
	MetricsPath string `yaml:"metrics_path" json:"metrics_path"`

	TlsConfig *tls.Config `yaml:"tls_config" json:"tls_config"`
}

type JMXBeans struct {
	Beans []map[string]interface{} `json:"beans"`
}

// SampleConfig returns the sample config
func (*HTTPMetricCollector) SampleConfig() string {
	return sampleConfig
}

// Description returns the description
func (*HTTPMetricCollector) Description() string {
	return ``
}

// Gather ...
func (p *HTTPMetricCollector) Gather() ([]*dto.MetricFamily, error) {

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
	defer func() {
		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
	}()

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

	var kept = JMXBeans{}
	if err := json.Unmarshal(bs, &kept); err != nil {
		return nil, err
	}

	metricFamilies := make(map[string]*dto.MetricFamily)

	for _, values := range kept.Beans {
		if len(values) == 0 {
			continue
		}

		nameValue, exist := values["name"]
		if !exist {
			continue
		}

		nameValueStr, ok := nameValue.(string)
		if !ok {
			continue
		}

		names := strings.Split(nameValueStr, ",name=")
		if len(names) <= 1 {
			continue
		}

		var metricPrefixes []string

		serviceStr := names[0]
		if ok := strings.Contains(serviceStr, ":service="); ok {
			services := strings.Split(serviceStr, ":service=")
			if len(services) <= 1 {
				continue
			}
			metricPrefixes = append(metricPrefixes, services[0], services[1])
		}

		namesSubs := strings.Split(names[1], ",sub=")
		name := strings.ToLower(namesSubs[0])

		metricPrefixes = append(metricPrefixes, name)
		if len(namesSubs) > 1 {
			metricPrefixes = append(metricPrefixes, namesSubs[1])
		}
		metricPrefix := strings.ToLower(strings.Join(metricPrefixes, "_"))

		tags := make(map[string]string)
		for key, value := range values {
			if value == nil || !strings.HasPrefix(key, "tag.") {
				continue
			}
			switch t := value.(type) {
			case string:
				tags[strings.TrimLeft(key, "tag.")] = t
			case int, int64, int32:
				tags[strings.TrimLeft(key, "tag.")] = fmt.Sprintf("%d", t)
			}
		}

		for key, value := range values {
			metricName := key
			metricName = fmt.Sprintf("%s_%s", metricPrefix, key)

			metricName = strings.ReplaceAll(metricName, " ", "_")
			metricName = strings.ReplaceAll(metricName, ".", "_")

			switch x := value.(type) {
			case bool:
				if x {
					value = 0
				} else {
					value = 1
				}
			case int, int32, int64, float32, float64:
			default:
				level.Warn(p.logger).Log("msg", "value is not numeric", key, value)
				continue
			}

			mf, ok := metricFamilies[metricName]
			if !ok {
				typ := dto.MetricType_UNTYPED
				mf = &dto.MetricFamily{
					Name: &metricName,
					Type: &typ,
				}
			}
			th, _ := types.ToFloat64(value)

			metric := &dto.Metric{
				Untyped: &dto.Untyped{
					Value: &th,
				},
			}
			for k, v := range tags {
				metric.Label = append(metric.Label, &dto.LabelPair{Name: &k, Value: &v})
			}

			mf.Metric = append(mf.GetMetric(), metric)
			metricFamilies[metricName] = mf
		}
	}

	var metrics []*dto.MetricFamily
	for _, mf := range metricFamilies {
		for i, metric := range mf.GetMetric() {
			metric.Label = append(metric.Label, &dto.LabelPair{Name: &instanceLabel, Value: &urlP.Host})
			mf.GetMetric()[i] = metric
		}
		metrics = append(metrics, mf)
	}
	return metrics, nil
}

func init() {
	inputs.RegisterFactory("http_jmx", func(opts ...plugins.Option) (plugins.InputMetricsCollector, error) {

		options := &plugins.Options{}
		for _, o := range opts {
			o(options)
		}

		p := &HTTPMetricCollector{
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

		return p, nil
	})
}
