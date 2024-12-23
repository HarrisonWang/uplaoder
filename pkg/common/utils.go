package common

import (
	"os"

	"gopkg.in/yaml.v3"
)

// LoadConfig 从配置文件加载配置
func LoadConfig(path string, config interface{}) error {
	// 如果配置文件存在，则读取
	if _, err := os.Stat(path); err == nil {
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		return yaml.Unmarshal(data, config)
	}
	return nil
}

// GetEnv 从环境变量获取配置项，若未定义则使用默认值
func GetEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}
