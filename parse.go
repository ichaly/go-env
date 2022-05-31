package env

import (
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"
)

type parseConfig struct {
	IgnorePrefix bool
}

type ParseOption func(*parseConfig)

func WithIgnorePrefix(ignorePrefix bool) ParseOption {
	return func(o *parseConfig) {
		o.IgnorePrefix = ignorePrefix
	}
}

type tag struct {
	Keys     []string
	Default  string
	Required bool
}

func parseTag(tagString string) tag {
	var t tag
	envKeys := strings.Split(tagString, ",")
	for _, key := range envKeys {
		if strings.Contains(key, "=") {
			keyData := strings.SplitN(key, "=", 2)
			switch strings.ToLower(keyData[0]) {
			case "default":
				t.Default = keyData[1]
			case "required":
				t.Required = strings.ToLower(keyData[1]) == "true"
			default:
				// just ignoring unsupported keys
				continue
			}
		} else {
			t.Keys = append(t.Keys, key)
		}
	}
	return t
}

//https://github.com/timest/env.git
//https://github.com/Netflix/go-env.git
//https://github.com/sethvargo/go-envconfig.git
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
	var prefixes []string
	if !cfg.IgnorePrefix {
		prefixes = append(prefixes, strings.ToUpper(rv.Type().Name()))
	}
	return unmarshal(prefixes, rv)
}

func unmarshal(prefixes []string, rv reflect.Value) error {
	for i := 0; i < rv.NumField(); i++ {
		fieldValue := rv.Field(i)
		if !fieldValue.CanSet() {
			return fmt.Errorf("field `%v`must be exported", fieldValue.String())
		}

		fieldType := rv.Type().Field(i)
		fieldKind := fieldValue.Kind()
		switch fieldKind {
		case reflect.Struct, reflect.Ptr:
			reflectValue := fieldValue
			if fieldKind == reflect.Ptr {
				reflectValue = reflect.New(fieldValue.Type().Elem())
				fieldValue.Set(reflectValue)
				reflectValue = reflectValue.Elem()
			}
			if err := unmarshal(append(prefixes, strings.ToUpper(fieldType.Name)), reflectValue); err != nil {
				return err
			}
		default:
			tagString := fieldType.Tag.Get("env")
			if len(tagString) == 0 {
				tagString = strings.ToUpper(fieldType.Name)
			}
			tag := parseTag(tagString)

			var ok bool
			var key string
			var value string
			for _, k := range tag.Keys {
				key = strings.Join(append(prefixes, strings.ToUpper(k)), "_")
				value, ok = os.LookupEnv(key)
				if ok {
					break
				}
			}
			if !ok && tag.Required {
				return fmt.Errorf("%s is required, but has not been set", key)
			}
			if len(value) == 0 {
				value = tag.Default
			}
			if err := fill(fieldValue, value); err != nil {
				return err
			}
		}
	}
	return nil
}

func fill(reflectValue reflect.Value, stringValue string) error {
	reflectType := reflectValue.Type()
	switch reflectValue.Kind() {
	case reflect.Bool:
		v, err := parseBool(stringValue)
		if err != nil {
			return err
		}
		reflectValue.SetBool(v)
	case reflect.String:
		reflectValue.SetString(stringValue)
	case reflect.Float32, reflect.Float64:
		v, err := strconv.ParseFloat(stringValue, reflectType.Bits())
		if err != nil {
			return err
		}
		reflectValue.SetFloat(v)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		// 如果是时长字段
		if reflectType.String() == "time.Duration" {
			if duration, err := time.ParseDuration(stringValue); err != nil {
				return err
			} else {
				reflectValue.Set(reflect.ValueOf(duration))
			}
			break
		}
		v, err := strconv.Atoi(stringValue)
		if err != nil {
			return err
		}
		reflectValue.SetInt(int64(v))
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		v, err := strconv.ParseUint(stringValue, 0, reflectType.Bits())
		if err != nil {
			return err
		}
		reflectValue.SetUint(v)
	case reflect.Slice:
		vals := strings.Split(stringValue, ",")
		s := reflect.MakeSlice(reflectType, len(vals), len(vals))
		for i, v := range vals {
			v = strings.TrimSpace(v)
			err := fill(s.Index(i), v)
			if err != nil {
				return err
			}
		}
		reflectValue.Set(s)
	case reflect.Map:
		vals := strings.Split(stringValue, ";")
		m := reflect.MakeMapWithSize(reflectType, len(vals))
		for _, val := range vals {
			pair := strings.SplitN(val, "=", 2)
			if len(pair) < 2 {
				return fmt.Errorf("invalid map item")
			}
			mKey, mVal := strings.TrimSpace(pair[0]), strings.TrimSpace(pair[1])

			k := reflect.New(reflectType.Key()).Elem()
			if err := fill(k, mKey); err != nil {
				return err
			}

			v := reflect.New(reflectType.Elem()).Elem()
			if err := fill(v, mVal); err != nil {
				return err
			}

			m.SetMapIndex(k, v)
		}
		reflectValue.Set(m)
	}
	return nil
}

func parseBool(v string) (bool, error) {
	if v == "" {
		return false, nil
	}
	return strconv.ParseBool(v)
}
