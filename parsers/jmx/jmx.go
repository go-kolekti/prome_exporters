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

package jmx

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	dto "github.com/prometheus/client_model/go"

	"trellis.tech/kolekti/prome_exporters/parsers"
	"trellis.tech/trellis/common.v1/types"
)

var (
	metricReg = regexp.MustCompile(`[\.|\-|\ ]`)
)

type Parser struct {
	logger log.Logger

	IgnoreTimestamp bool
}

func NewParser(logger log.Logger) (parsers.Parser, error) {
	return &Parser{logger: logger}, nil
}

type beans struct {
	Beans []map[string]interface{} `json:"beans"`
}

func (p *Parser) Parse(bs []byte, tags map[string]string, _ string) (map[string]*dto.MetricFamily, error) {

	var kept = beans{}
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

			metricName = metricReg.ReplaceAllString(metricName, "_")

			switch x := value.(type) {
			case bool:
				if x {
					value = 0
				} else {
					value = 1
				}
			case int, int32, int64, float32, float64:
			default:
				_ = level.Warn(p.logger).Log("msg", "value is not numeric", key, value)
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
				key, value := k, v
				metric.Label = append(metric.Label, &dto.LabelPair{Name: &key, Value: &value})
			}

			mf.Metric = append(mf.GetMetric(), metric)
			metricFamilies[metricName] = mf
		}
	}

	return metricFamilies, nil
}
