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

var defaultSerializerConfig = &SerializerConfig{
	Name: "prometheus",
}

type SerializerConfig struct {
	Name string `yaml:"name" json:"name"`
}

func NewSerializer(c *SerializerConfig) (Serializer, error) {
	if c == nil || c.Name == "" {
		c = defaultSerializerConfig
	}

	f, err := GetFactory(c.Name)
	if err != nil {
		return nil, err
	}

	return f()
}
