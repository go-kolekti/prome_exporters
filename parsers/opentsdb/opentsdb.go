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

package opentsdb

import (
	"encoding/json"
	"regexp"
	"strconv"
	"time"

	"github.com/go-kit/log"
	dto "github.com/prometheus/client_model/go"
	"trellis.tech/kolekti/prome_exporters/parsers"
)

var (
	metricReg = regexp.MustCompile("[.-]")
)

type Parser struct {
	logger log.Logger
	cfg    parsers.Config
}

type metrics []struct {
	Metric    string            `json:"metric"`
	Timestamp int64             `json:"timestamp"`
	Value     json.Number       `json:"value"`
	Tags      map[string]string `json:"tags"`
}

func NewParser(logger log.Logger, cfg parsers.Config) (parsers.Parser, error) {
	return &Parser{logger: logger, cfg: cfg}, nil
}

func (p *Parser) Parse(bs []byte, tags map[string]string, _ string) (map[string]*dto.MetricFamily, error) {

	ms := metrics{}

	if err := json.Unmarshal(bs, &ms); err != nil {
		return nil, err
	}

	metricFamilies := make(map[string]*dto.MetricFamily)

	for _, m := range ms {
		metricName := metricReg.ReplaceAllString(m.Metric, "_")
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
		if !p.cfg.JMXOptions.IgnoreTimestamp {
			timestampMs := timestampMsFunc(m.Timestamp)
			metric.TimestampMs = &timestampMs
		}

		for k, v := range m.Tags {
			key, value := k, v
			metric.Label = append(metric.Label, &dto.LabelPair{Name: &key, Value: &value})
		}

		for k, v := range tags {
			key, value := k, v //
			metric.Label = append(metric.Label, &dto.LabelPair{Name: &key, Value: &value})
		}

		mf.Metric = append(mf.Metric, metric)

		metricFamilies[metricName] = mf
	}

	return metricFamilies, nil
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
