package configs

import (
	"os"

	"github.com/harrisonwang/media-processor/pkg/common"
)

type Config struct {
	Server struct {
		Port string `yaml:"port"`
	} `yaml:"server"`
	Upload struct {
		Path           string `yaml:"path"`
		MediaUrlPrefix string `yaml:"media_url_prefix"`
	} `yaml:"upload"`
	OCR struct {
		Endpoint                    string `yaml:"endpoint"`
		AlibabaCloudAccessKeyID     string `yaml:"alibaba_cloud_access_key_id"`
		AlibabaCloudAccessKeySecret string `yaml:"alibaba_cloud_access_key_secret"`
	} `yaml:"ocr"`
}

func Load() *Config {
	config := &Config{}

	// 尝试从配置文件加载
	err := common.LoadConfig("configs/config.yaml", config)
	if err != nil {
		// 记录错误但继续执行，因为还可以从环境变量加载
		println("Warning: Failed to load config file:", err.Error())
	}

	// 如果环境变量存在，则覆盖配置文件的值
	if val := os.Getenv("UPLOAD_PATH"); val != "" {
		config.Upload.Path = val
	}
	if val := os.Getenv("MEDIA_URL_PREFIX"); val != "" {
		config.Upload.MediaUrlPrefix = val
	}
	if val := os.Getenv("OCR_ENDPOINT"); val != "" {
		config.OCR.Endpoint = val
	}
	if val := os.Getenv("ALIBABA_CLOUD_ACCESS_KEY_ID"); val != "" {
		config.OCR.AlibabaCloudAccessKeyID = val
	}
	if val := os.Getenv("ALIBABA_CLOUD_ACCESS_KEY_SECRET"); val != "" {
		config.OCR.AlibabaCloudAccessKeySecret = val
	}
	if val := os.Getenv("SERVER_PORT"); val != "" {
		config.Server.Port = val
	}

	return config
}
