package env

import (
	"errors"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"
)

var (
	compileRegex = regexp.MustCompile(`\$\{.*?}`)
)

type loadConfig struct {
	StrictMode bool
}

type LoadOption func(*loadConfig)

func WithStrictMode(strictMode bool) LoadOption {
	return func(o *loadConfig) {
		o.StrictMode = strictMode
	}
}

func String(s string, opts ...LoadOption) (res string, err error) {
	// 设置配置
	cfg := loadConfig{
		StrictMode: false,
	}
	for _, o := range opts {
		o(&cfg)
	}
	// 解析所有的环境变量到map中
	envMap := make(map[string]interface{})
	for _, v := range os.Environ() {
		splitString := strings.SplitN(v, "=", 2)
		envMap[splitString[0]] = splitString[1]
	}
	// 初始化返回结果
	res = s
	// 匹配所有到变量
	vars := compileRegex.FindAllString(s, -1)
	for _, v := range vars {
		// 去掉头部的`${`和尾部的`}`
		key := v[2 : len(v)-1]
		value := ""
		// 如果有默认值进行字符串拆分
		if strings.Contains(key, ":=") {
			arr := strings.Split(key, ":=")
			key = arr[0]
			value = arr[1]
		}
		// 格式化变量名为大写并去除多余空格
		key = strings.ToUpper(strings.TrimSpace(key))
		val, ok := os.LookupEnv(key)
		if ok {
			value = val
		} else if len(value) == 0 && cfg.StrictMode {
			return "", errors.New(fmt.Sprintf("Key `%v` is not set", key))
		}
		log.Printf("----->>>%v:%v", key, value)
		// 使用获取到的变量值替换原字符串中的占位符
		res = strings.ReplaceAll(res, v, strings.TrimSpace(value))
	}
	return
}
