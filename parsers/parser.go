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

package parsers

import (
	"regexp"

	dto "github.com/prometheus/client_model/go"
)

type Option func(*Config)

type Config struct {
	Name string `yaml:"name" json:"name"`

	PrefixWhitelist []string `yaml:"prefix_whitelist" json:"prefix_whitelist"`
	PrefixBlacklist []string `yaml:"prefix_blacklist" json:"prefix_blacklist"`

	Whitelists []*regexp.Regexp `yaml:"-" json:"-"`
	Blacklists []*regexp.Regexp `yaml:"-" json:"-"`

	// Prometheus
	PrometheusOptions `yaml:",inline" json:",inline"`

	// JMX
	JMXOptions `yaml:",inline" json:",inline"`
}

type PrometheusOptions struct {
}

type JMXOptions struct {
	IgnorePrefix    bool `yaml:"jmx_ignore_prefix" json:"jmx_ignore_prefix"`
	IgnoreTimestamp bool `yaml:"jmx_ignore_timestamp" json:"jmx_ignore_timestamp"`
}

type Parser interface {
	Parse(bs []byte, tags map[string]string, ct string) (map[string]*dto.MetricFamily, error)
}
