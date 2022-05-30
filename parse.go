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
	return unmarshal(prefixes, rv, cfg)
}

func unmarshal(prefixes []string, rv reflect.Value, cfg parseConfig) error {
	for i := 0; i < rv.NumField(); i++ {
		valueField := rv.Field(i)
		if !valueField.CanSet() {
			return fmt.Errorf("field must be exported")
		}

		typeField := rv.Type().Field(i)
		switch valueField.Kind() {
		case reflect.Struct:
			err := unmarshal(append(prefixes, strings.ToUpper(typeField.Name)), valueField.Field(i), cfg)
			if err != nil {
				return err
			}
		default:
			tagString := typeField.Tag.Get("env")
			if len(tagString) == 0 {
				tagString = strings.ToUpper(typeField.Name)
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
			if err := fill(valueField, value); err != nil {
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
		if v, err := parseBool(stringValue); err != nil {
			return err
		} else {
			reflectValue.SetBool(v)
		}
	case reflect.String:
		reflectValue.SetString(stringValue)
	case reflect.Float32, reflect.Float64:
		if v, err := strconv.ParseFloat(stringValue, reflectType.Bits()); err != nil {
			return err
		} else {
			reflectValue.SetFloat(v)
		}
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
		if v, err := strconv.Atoi(stringValue); err != nil {
			return err
		} else {
			reflectValue.SetInt(int64(v))
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		if v, err := strconv.ParseUint(stringValue, 0, reflectType.Bits()); err != nil {
			return err
		} else {
			reflectValue.SetUint(v)
		}
	case reflect.Ptr:
		t := reflectValue.Type()
		ptr := reflect.New(t.Elem())
		if err := fill(ptr.Elem(), stringValue); err != nil {
			return err
		} else {
			reflectValue.Set(ptr)
		}
	case reflect.Map:
	case reflect.Slice:
	}
	return nil
}

func parseBool(v string) (bool, error) {
	if v == "" {
		return false, nil
	}
	return strconv.ParseBool(v)
}

func parse(prefix string, f reflect.Value, sf reflect.StructField) error {
	df := sf.Tag.Get("default")
	isRequire, err := parseBool(sf.Tag.Get("require"))
	if err != nil {
		return fmt.Errorf("the value of %s is not a valid `member` of bool ，only "+
			"[1 0 t f T F true false TRUE FALSE True False] are supported", prefix)
	}
	ev, exist := os.LookupEnv(prefix)

	if !exist && isRequire {
		return fmt.Errorf("%s is required, but has not been set", prefix)
	}
	if !exist && df != "" {
		ev = df
	}
	//log.Print("ev:", ev)
	switch f.Kind() {
	case reflect.String:
		f.SetString(ev)
	case reflect.Int:
		iv, err := strconv.ParseInt(ev, 10, 32)
		if err != nil {
			return fmt.Errorf("%s:%s", prefix, err)
		}
		f.SetInt(iv)
	case reflect.Int64:
		if f.Type().String() == "time.Duration" {
			t, err := time.ParseDuration(ev)
			if err != nil {
				return fmt.Errorf("%s:%s", prefix, err)
			}
			f.Set(reflect.ValueOf(t))
		} else {
			iv, err := strconv.ParseInt(ev, 10, 64)
			if err != nil {
				return fmt.Errorf("%s:%s", prefix, err)
			}
			f.SetInt(iv)
		}
	case reflect.Uint:
		uiv, err := strconv.ParseUint(ev, 10, 32)
		if err != nil {
			return fmt.Errorf("%s:%s", prefix, err)
		}
		f.SetUint(uiv)
	case reflect.Uint64:
		uiv, err := strconv.ParseUint(ev, 10, 64)
		if err != nil {
			return fmt.Errorf("%s:%s", prefix, err)
		}
		f.SetUint(uiv)
	case reflect.Float32:
		f32, err := strconv.ParseFloat(ev, 32)
		if err != nil {
			return fmt.Errorf("%s:%s", prefix, err)
		}
		f.SetFloat(f32)
	case reflect.Float64:
		f64, err := strconv.ParseFloat(ev, 64)
		if err != nil {
			return fmt.Errorf("%s:%s", prefix, err)
		}
		f.SetFloat(f64)
	case reflect.Bool:
		b, err := parseBool(ev)
		if err != nil {
			return fmt.Errorf("%s:%s", prefix, err)
		}
		f.SetBool(b)
	case reflect.Slice:
		sep := ";"
		s, exist := sf.Tag.Lookup("slice_sep")
		if exist && s != "" {
			sep = s
		}
		vals := strings.Split(ev, sep)
		switch f.Type() {
		case reflect.TypeOf([]string{}):
			f.Set(reflect.ValueOf(vals))
		case reflect.TypeOf([]int{}):
			t := make([]int, len(vals))
			for i, v := range vals {
				val, err := strconv.ParseInt(v, 10, 32)
				if err != nil {
					return fmt.Errorf("%s:%s", prefix, err)
				}
				t[i] = int(val)
			}
		case reflect.TypeOf([]int64{}):
			t := make([]int64, len(vals))
			for i, v := range vals {
				val, err := strconv.ParseInt(v, 10, 64)
				if err != nil {
					return fmt.Errorf("%s:%s", prefix, err)
				}
				t[i] = val
			}
		case reflect.TypeOf([]uint{}):
			t := make([]uint, len(vals))
			for i, v := range vals {
				val, err := strconv.ParseUint(v, 10, 32)
				if err != nil {
					return fmt.Errorf("%s:%s", prefix, err)
				}
				t[i] = uint(val)
			}
		case reflect.TypeOf([]uint64{}):
			t := make([]uint64, len(vals))
			for i, v := range vals {
				val, err := strconv.ParseUint(v, 10, 64)
				if err != nil {
					return fmt.Errorf("%s:%s", prefix, err)
				}
				t[i] = val
			}
		case reflect.TypeOf([]float32{}):
			t := make([]float32, len(vals))
			for i, v := range vals {
				val, err := strconv.ParseFloat(v, 32)
				if err != nil {
					return fmt.Errorf("%s:%s", prefix, err)
				}
				t[i] = float32(val)
			}
		case reflect.TypeOf([]float64{}):
			t := make([]float64, len(vals))
			for i, v := range vals {
				val, err := strconv.ParseFloat(v, 64)
				if err != nil {
					return fmt.Errorf("%s:%s", prefix, err)
				}
				t[i] = val
			}
		case reflect.TypeOf([]bool{}):
			t := make([]bool, len(vals))
			for i, v := range vals {
				val, err := parseBool(v)
				if err != nil {
					return fmt.Errorf("%s:%s", prefix, err)
				}
				t[i] = val
			}
		}
	}
	return nil
}
