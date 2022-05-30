package env

import (
	"fmt"
	"reflect"
	"strings"
)

type parseConfig struct {
	IgnorePrefix bool
}

type ParseOption func(*parseConfig)

func WithIgnorePrefix(strictMode bool) ParseOption {
	return func(o *parseConfig) {
		o.IgnorePrefix = strictMode
	}
}

func Parse(v interface{}, opts ...ParseOption) error {
	// 校验参数
	rv := reflect.Indirect(reflect.ValueOf(v))
	if reflect.ValueOf(v).Kind() != reflect.Ptr || rv.Kind() != reflect.Struct {
		return fmt.Errorf("only the pointer to a struct is supported")
	}
	// 设置配置
	cfg := parseConfig{
		IgnorePrefix: false,
	}
	for _, o := range opts {
		o(&cfg)
	}
	prefix := strings.ToUpper(rv.Type().Name())
	if cfg.IgnorePrefix {
		prefix = ""
	}
	return fill(prefix, rv)
}

func fill(prefix string, rv reflect.Value) error {
	return nil
}
