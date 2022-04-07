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

package inputs

import (
	"reflect"
	"strings"

	"trellis.tech/kolekti/prome_exporters/plugins"

	"trellis.tech/trellis/common.v1/errcode"
)

type FactoryPrometheusCollector func(...plugins.Option) (plugins.InputPrometheusCollector, error)
type FactoryMetricsCollector func(...plugins.Option) (plugins.InputMetricsCollector, error)

var (
	mapNewCollectorFunc = make(map[string]*Input)
)

type Input struct {
	inputType plugins.InputType

	pcFactory FactoryPrometheusCollector
	mcFactory FactoryMetricsCollector
}

func (p *Input) InputType() plugins.InputType {
	return p.inputType
}

func (p *Input) NewPrometheusCollector(opts ...plugins.Option) (plugins.InputPrometheusCollector, error) {
	if p.inputType != plugins.InputTypePrometheusCollector || p.pcFactory == nil {
		return nil, errcode.Newf("its not prometheus collector factory, type: %d, %+v", p.inputType, p.pcFactory)
	}

	return p.pcFactory(opts...)
}

func (p *Input) NewMetricsCollector(opts ...plugins.Option) (plugins.InputMetricsCollector, error) {
	if p.inputType != plugins.InputTypeMetricsCollector || p.mcFactory == nil {
		return nil, errcode.Newf("its not metrics collector factory, type: %d, %+v", p.inputType, p.mcFactory)
	}

	return p.mcFactory(opts...)
}

func RegisterFactory(name string, fn interface{}) {
	if name = strings.TrimSpace(name); name == "" {
		panic(errcode.New("empty collector name"))
	}
	if fn == nil {
		panic(errcode.New("nil collector factory"))
	}
	if _, ok := mapNewCollectorFunc[name]; ok {
		panic(errcode.Newf("collector factory already exists: %s", name))
	}

	var input = &Input{}
	switch t := fn.(type) {
	case FactoryPrometheusCollector:
		input.inputType = plugins.InputTypePrometheusCollector
		input.pcFactory = t
	case func(...plugins.Option) (plugins.InputPrometheusCollector, error):
		input.inputType = plugins.InputTypePrometheusCollector
		input.pcFactory = t
	case FactoryMetricsCollector:
		input.inputType = plugins.InputTypeMetricsCollector
		input.mcFactory = t
	case func(...plugins.Option) (plugins.InputMetricsCollector, error):
		input.inputType = plugins.InputTypeMetricsCollector
		input.mcFactory = t

	default:
		panic(errcode.Newf("not supported type : %s, %+v", name, reflect.TypeOf(t).String()))
	}
	mapNewCollectorFunc[name] = input
}

func GetFactory(name string) (*Input, error) {
	input, ok := mapNewCollectorFunc[name]
	if !ok || input == nil {
		return nil, errcode.Newf("not found collector factory: %s", name)
	}
	return input, nil
}
