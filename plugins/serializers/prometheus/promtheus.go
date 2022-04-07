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
	"bytes"

	"trellis.tech/kolekti/prome_exporters/plugins/serializers"

	dto "github.com/prometheus/client_model/go"
	"github.com/prometheus/common/expfmt"
)

func init() {
	serializers.RegisterFactory("prometheus", NewSerializer)
}

type Serializer struct{}

func NewSerializer(...serializers.Option) (serializers.Serializer, error) {
	return &Serializer{}, nil
}

func (s *Serializer) Serialize(metric *dto.MetricFamily) ([]byte, error) {
	return s.SerializeBatch([]*dto.MetricFamily{metric})
}

func (s *Serializer) SerializeBatch(metrics []*dto.MetricFamily) ([]byte, error) {

	var buf bytes.Buffer
	for _, mf := range metrics {
		enc := expfmt.NewEncoder(&buf, expfmt.FmtText)
		err := enc.Encode(mf)
		if err != nil {
			return nil, err
		}
	}

	return buf.Bytes(), nil
}
