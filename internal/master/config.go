package master

import (
	"github.com/caarlos0/env/v10"
)

var Cfg = LoadConfig()

type Config struct {
	Configpath string `env:"CONFIG_PATH"`
}

func LoadConfig() *Config {
	cfg := &Config{}
	if err := env.Parse(cfg); err != nil {
		log.Fatalf("环境变量解析失败: %v", err)
	}
	if cfg.Configpath == "" {
		log.Fatalf("❌ CONFIG_PATH 是必需的环境变量，当前未设置")
	}
	return cfg
}
