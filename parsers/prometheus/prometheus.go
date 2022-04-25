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

package prometheus

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"mime"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/matttproud/golang_protobuf_extensions/pbutil"
	dto "github.com/prometheus/client_model/go"
	"github.com/prometheus/common/expfmt"
	"trellis.tech/kolekti/prome_exporters/parsers"
)

type Parser struct {
	cfg    parsers.Config
	logger log.Logger
}

func NewParser(logger log.Logger, cfg parsers.Config) (parsers.Parser, error) {
	return &Parser{logger: logger, cfg: cfg}, nil
}

func (p *Parser) Parse(body []byte, tags map[string]string, contentType string) (map[string]*dto.MetricFamily, error) {

	var parser expfmt.TextParser
	// Read raw data
	buffer := bytes.NewBuffer(body)
	reader := bufio.NewReader(buffer)

	metricFamilies := make(map[string]*dto.MetricFamily)

	mediatype, params, err := mime.ParseMediaType(contentType)
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
			return nil, err
		}
	}

	for k, v := range tags {
		key, value := k, v
		for _, family := range metricFamilies {
			for _, metric := range family.GetMetric() {
				metric.Label = append(metric.Label, &dto.LabelPair{Name: &key, Value: &value})
			}
		}
	}

	return metricFamilies, nil
}
