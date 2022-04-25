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
	"reflect"
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
	cfg    parsers.Config
}

func NewParser(logger log.Logger, cfg parsers.Config) (parsers.Parser, error) {
	p := &Parser{
		logger: logger,
		cfg:    cfg,
	}

	return p, nil
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

	metricFiltered := false
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

		if !p.cfg.JMXOptions.IgnorePrefix {
			serviceStr := names[0]
			if ok := strings.Contains(serviceStr, ":service="); ok {
				services := strings.Split(serviceStr, ":service=")
				if len(services) <= 1 {
					continue
				}
				metricPrefixes = append(metricPrefixes, strings.TrimSpace(services[0]), strings.TrimSpace(services[1]))
			}
		}

		namesSubs := strings.Split(names[1], ",sub=")
		tags["name"] = strings.TrimSpace(namesSubs[0])

		if len(namesSubs) > 1 {
			tags["sub"] = strings.TrimSpace(namesSubs[1])
		} else {
			delete(tags, "sub")
		}

		for key, value := range values {
			if value == nil || !strings.HasPrefix(key, "tag.") {
				continue
			}
			switch t := value.(type) {
			case string:
				if t = strings.TrimSpace(t); t != "" {
					tags[strings.TrimLeft(key, "tag.")] = t
				}
			case int, int64, int32:
				tags[strings.TrimLeft(key, "tag.")] = fmt.Sprintf("%d", t)
			case float32, float64:
				tags[strings.TrimLeft(key, "tag.")] = fmt.Sprintf("%f", t)
			}
		}

		for key, value := range values {
			if value == nil {
				continue
			}
			//metricName := key
			metricNames := append(metricPrefixes, strings.TrimSpace(key))
			metricName := strings.Join(metricNames, "_")

			metricName = metricReg.ReplaceAllString(metricName, "_")

			for _, wl := range p.cfg.Whitelists {
				if wl.MatchString(metricName) {
					// 发现匹配白名单字符串，不被过滤
					metricFiltered = false
					break
				} else {
					// 发现不匹配白名单字符串，过滤
					metricFiltered = true
				}
			}

			if metricFiltered {
				level.Debug(p.logger).Log("msg", "filter metric with whitelist", "metric", metricName)
				continue
			}

			for _, bl := range p.cfg.Blacklists {
				if bl.MatchString(metricName) {
					// 发现匹配黑名单字符串，被过滤
					metricFiltered = true
					break
				} else {
					// 发现不匹配黑名单字符串，过滤
					metricFiltered = false
				}
			}

			if metricFiltered {
				level.Debug(p.logger).Log("msg", "filter metric with blacklist", "metric", metricName)
				continue
			}

			var th = 0.0
			switch x := reflect.TypeOf(value).Kind(); x {
			case reflect.Bool:
				if value == true {
					th = 1
				} else {
					th = 0
				}
			default:
				tValue, err := types.ToFloat64(value)
				if err != nil {
					_ = level.Warn(p.logger).Log("msg", "value is not numeric", key, value)
					continue
				}
				th = tValue
			}

			mf, ok := metricFamilies[metricName]
			if !ok {
				typ := dto.MetricType_UNTYPED
				mf = &dto.MetricFamily{
					Name: &metricName,
					Type: &typ,
				}
			}

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
