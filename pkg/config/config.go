package config

import (
	"log"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server        ServerConfig        `yaml:"server"`
	Logger        LoggerConfig        `yaml:"logger"`
	AliyunOSS     AliyunOSSConfig     `yaml:"aliyun_oss"`
	MediaTransfer MediaTransferConfig `yaml:"media_transfer"`
}

type ServerConfig struct {
	Port   int    `yaml:"port"`
	Mode   string `yaml:"mode"`
	Domain string `yaml:"domain"`
}

type LoggerConfig struct {
	Level  string     `yaml:"level"`
	Output string     `yaml:"output"`
	File   FileConfig `yaml:"file"`
}

type FileConfig struct {
	Path       string `yaml:"path"`
	MaxSize    int    `yaml:"max_size"`
	MaxAge     int    `yaml:"max_age"`
	MaxBackups int    `yaml:"max_backups"`
	Compress   bool   `yaml:"compress"`
}

type AliyunOSSConfig struct {
	AccessKeyID     string `yaml:"access_key_id"`
	AccessKeySecret string `yaml:"access_key_secret"`
	Endpoint        string `yaml:"endpoint"`
	Bucket          string `yaml:"bucket"`
	Region          string `yaml:"region"`
}

type MediaTransferConfig struct {
	DownloadTimeout int64    `yaml:"download_timeout"`
	RetryCount      int      `yaml:"retry_count"`
	AllowedDomains  []string `yaml:"allowed_domains"`
}

var GlobalConfig Config

func Init() {
	// 获取配置文件路径
	configPath := os.Getenv("TRANSFER_CONFIG")
	if configPath == "" {
		configPath = "local.yaml"
	}

	pwd, err := os.Getwd()
	if err != nil {
		log.Printf("获取当前工作目录失败: %v", err)
	} else {
		log.Printf("当前工作目录: %s", pwd)
	}

	log.Printf("读取配置文件: %s", pwd+"/config/"+configPath)
	configFile, err := os.ReadFile("config/" + configPath)
	if err != nil {
		log.Fatalf("读取配置文件失败: %v", err)
	}

	if err := yaml.Unmarshal(configFile, &GlobalConfig); err != nil {
		log.Fatalf("解析配置文件失败: %v", err)
	}

	log.Printf("配置加载成功")
}
