package outputs

import (
	"strings"

	"trellis.tech/trellis/common.v1/errcode"
)

var outputs = map[string]Factory{}

func RegisterFactory(name string, fn interface{}) {
	if name = strings.TrimSpace(name); name == "" {
		panic(errcode.New("empty output name"))
	}
	if fn == nil {
		panic(errcode.New("nil collector factory"))
	}

	switch f := fn.(type) {
	case Factory:
		outputs[name] = f
	case func(...Option) (Output, error):
		outputs[name] = f
	default:
		panic(errcode.Newf("not supported output factory: %s", name))
	}
}

func GetFactory(name string) (Factory, error) {
	fn, ok := outputs[name]
	if !ok || fn == nil {
		return nil, errcode.Newf("not found output factory: %s", name)
	}
	return fn, nil
}