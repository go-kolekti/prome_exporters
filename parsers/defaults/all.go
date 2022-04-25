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
	"regexp"

	"trellis.tech/kolekti/prome_exporters/parsers"
	"trellis.tech/kolekti/prome_exporters/parsers/jmx"
	"trellis.tech/kolekti/prome_exporters/parsers/opentsdb"
	"trellis.tech/kolekti/prome_exporters/parsers/prometheus"

	"github.com/go-kit/log"
)

func NewParser(logger log.Logger, cfg parsers.Config) (parsers.Parser, error) {

	for _, s := range cfg.PrefixWhitelist {
		cfg.Whitelists = append(cfg.Whitelists, regexp.MustCompile(s))
	}

	for _, s := range cfg.PrefixBlacklist {
		cfg.Blacklists = append(cfg.Blacklists, regexp.MustCompile(s))
	}

	switch cfg.Name {
	case "", "prometheus":
		return prometheus.NewParser(logger, cfg)
	case "jmx":
		return jmx.NewParser(logger, cfg)
	case "opentsdb":
		return opentsdb.NewParser(logger, cfg)
	default:
		return nil, fmt.Errorf("unsupported parser type")
	}
}
