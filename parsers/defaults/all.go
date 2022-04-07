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

package defaults

import (
	"fmt"

	"trellis.tech/kolekti/prome_exporters/parsers/jmx"

	"github.com/go-kit/log"
	"trellis.tech/kolekti/prome_exporters/parsers"
	"trellis.tech/kolekti/prome_exporters/parsers/opentsdb"
	"trellis.tech/kolekti/prome_exporters/parsers/prometheus"
)

func NewParser(name string, logger log.Logger) (parsers.Parser, error) {
	switch name {
	case "", "prometheus":
		return prometheus.NewParser(logger)
	case "jmx":
		return jmx.NewParser(logger)
	case "opentsdb":
		return opentsdb.NewParser(logger)
	default:
		return nil, fmt.Errorf("unsupported parser type")
	}
}
