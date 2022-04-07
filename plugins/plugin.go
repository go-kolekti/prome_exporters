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

package plugins

import (
	"github.com/go-kit/log"
	"trellis.tech/trellis/common.v1/config"
)

// PluginDescriber contains the functions defaults plugins must implement to describe
// themselves to Telegraf. Note that defaults plugins may define a logger that is
// not part of the interface, but will receive an injected logger if it's set.
// eg: Log telegraf.Logger `toml:"-"`
type PluginDescriber interface {
	// SampleConfig returns the defaults configuration of the Processor
	SampleConfig() string

	// Description returns a one-sentence description on the Processor
	Description() string
}

type Option func(*Options)
type Options struct {
	Config config.Config
	Logger log.Logger
}

func Config(c config.Config) Option {
	return func(o *Options) {
		o.Config = c
	}
}

func Logger(l log.Logger) Option {
	return func(o *Options) {
		o.Logger = l
	}
}
