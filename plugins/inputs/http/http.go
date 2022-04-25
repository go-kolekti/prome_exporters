/*
Copyright © 2022 Henry Huang <hhh@rutcode.com>
This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.
This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.
You should have received a copy of the GNU General Public License
along with this program. If not, see <http://www.gnu.org/licenses/>.
*/

package http

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	"trellis.tech/kolekti/prome_exporters/internal"
	"trellis.tech/kolekti/prome_exporters/parsers"
	"trellis.tech/kolekti/prome_exporters/parsers/defaults"
	"trellis.tech/kolekti/prome_exporters/plugins"
	"trellis.tech/kolekti/prome_exporters/plugins/inputs"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	dto "github.com/prometheus/client_model/go"
	"trellis.tech/trellis/common.v1/config"
	"trellis.tech/trellis/common.v1/crypto/tls"
	"trellis.tech/trellis/common.v1/types"
)

var (
	maxErrMsgLen int64 = 1024

	defaultTimeout = 10 * time.Second
	labelInstance  = "instance"
)

type Collector struct {
	client *http.Client
	logger log.Logger

	Timeout types.Duration    `yaml:"timeout" json:"timeout"`
	Urls    []string          `yaml:"urls" json:"urls"`
	Headers map[string]string `yaml:"headers" json:"headers"`

	TlsConfig *tls.Config `yaml:"tls_config" json:"tls_config"`

	Tags map[string]string `yaml:"tags" json:"tags"`

	Parser parsers.Config `yaml:"parser" json:"parser"`

	parser parsers.Parser
}

// SampleConfig returns the sample config
func (*Collector) SampleConfig() string {
	return ``
}

// Description returns the description
func (*Collector) Description() string {
	return ``
}

// Gather ...
func (p *Collector) Gather() ([]*dto.MetricFamily, error) {

	mfs := make(map[string]*dto.MetricFamily)
	for _, urlStr := range p.Urls {
		urlP, err := url.Parse(urlStr)
		if err != nil {
			level.Error(p.logger).Log("msg", "parse_url_failed", "url", urlStr, "error", err)
			continue
		}
		mfsServer, err := p.gatherServer(urlP)
		if err != nil {
			level.Error(p.logger).Log("msg", "gather_server_failed", "url", urlStr, "error", err)
			continue
		}
		for name, family := range mfsServer {
			mf, ok := mfs[name]
			if !ok {
				mfs[name] = family
				continue
			}
			mf.Metric = append(mf.Metric, family.GetMetric()...)
		}
	}

	var metrics []*dto.MetricFamily
	for _, family := range mfs {
		metrics = append(metrics, family)
	}
	return metrics, nil
}

func (p *Collector) gatherServer(urlP *url.URL) (map[string]*dto.MetricFamily, error) {

	req, _ := http.NewRequest("GET", urlP.String(), nil)
	for key, value := range p.Headers {
		req.Header.Set(key, value)
	}
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

	tags := make(map[string]string)
	if p.Tags != nil {
		tags = config.DeepCopy(p.Tags).(map[string]string)
	}

	tags[labelInstance] = urlP.Host

	return p.parser.Parse(bs, tags, resp.Header.Get("Content-Type"))
}

func init() {
	inputs.RegisterFactory("http", func(opts ...plugins.Option) (_ plugins.InputMetricsCollector, err error) {

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

		p.parser, err = defaults.NewParser(log.With(p.logger, "parser", p.Parser.Name), p.Parser)
		if err != nil {
			return nil, err
		}

		return p, nil
	})
}
