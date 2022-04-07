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

package promethues_node_exporter

import (
	"trellis.tech/kolekti/prome_exporters/internal"
	"trellis.tech/kolekti/prome_exporters/plugins"
	"trellis.tech/kolekti/prome_exporters/plugins/inputs"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/node_exporter/collector"
)

func init() {
	inputs.RegisterFactory("prometheus_node_exporter", NewNodeExporterCollector)
}

type Collector struct {
	prometheus.Collector
}

func (p *Collector) Tags() map[string]string {
	return map[string]string{
		"instance": internal.GetIP(),
	}
}

func NewNodeExporterCollector(opts ...plugins.Option) (_ plugins.InputPrometheusCollector, err error) {
	options := &plugins.Options{}
	for _, opt := range opts {
		opt(options)
	}

	var filters []string
	if options.Config != nil {
		filters = options.Config.GetStringList("filters")
	}

	c := &Collector{}
	c.Collector, err = collector.NewNodeCollector(options.Logger, filters...)
	if err != nil {
		return
	}

	return c, nil
}
