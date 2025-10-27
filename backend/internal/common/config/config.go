package config

import (
	"fmt"

	"github.com/spf13/viper"
)

type PostgresConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	SSLMode  string
}

type MongoConfig struct {
	URI      string
	Database string
}

type MinioConfig struct {
	Endpoint  string
	AccessKey string `mapstructure:"access_key"`
	SecretKey string `mapstructure:"secret_key"`
	Bucket    string
	UseSSL    bool `mapstructure:"use_ssl"`
}

type RedisConfig struct {
	Addr     string
	Password string
	DB       int
}

type Config struct {
	Server struct {
		Domain string
		Port   int
	}
	Postgres PostgresConfig
	MongoDB  MongoConfig
	Minio    MinioConfig
	Redis    RedisConfig
}

// 常量定义
var (
	// regex
	EmailPattern = `^[a-zA-Z0-9._%+-]+@hrbeu.edu.cn`
	ImagePattern = `(?i)\.(jpg|jpeg|png|bmp)$` // ?i 表示忽略大小写

	// SnowFlake
	MachineID = 0
)

// 加载yaml配置文件
func LoadConfig(path string) (Config, error) {
	var cfg Config
	viper.SetConfigFile(path)
	if err := viper.ReadInConfig(); err != nil {
		return cfg, fmt.Errorf("failed to read config file: %w", err)
	}
	if err := viper.Unmarshal(&cfg); err != nil {
		return cfg, fmt.Errorf("failed to load config: %w", err)
	}
	return cfg, nil
}
