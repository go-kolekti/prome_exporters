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

package serializers

import (
	"strings"

	"github.com/go-kit/log"
	"trellis.tech/trellis/common.v1/config"
	"trellis.tech/trellis/common.v1/errcode"
)

type Factory func(...Option) (Serializer, error)

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

var outputs = map[string]Factory{}

func RegisterFactory(name string, fn interface{}) {
	if name = strings.TrimSpace(name); name == "" {
		panic(errcode.New("empty serializer name"))
	}
	if fn == nil {
		panic(errcode.New("nil serializer factory"))
	}

	switch f := fn.(type) {
	case Factory:
		outputs[name] = f
	case func(...Option) (Serializer, error):
		outputs[name] = f
	default:
		panic(errcode.Newf("not supported serializer factory: %s", name))
	}
}

func GetFactory(name string) (Factory, error) {
	fn, ok := outputs[name]
	if !ok || fn == nil {
		return nil, errcode.Newf("not found serializer factory: %s", name)
	}
	return fn, nil
}
