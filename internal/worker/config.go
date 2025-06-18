package worker

import (
	"fmt"
	"github.com/caarlos0/env/v10"
	"time"
)

type Config struct {
	MasterServerAddr string     `env:"MASTER_SERVER_ADDR"`
	TimeStartRaw     string     `env:"TICKET_TIME_START"` // 原始字符串
	TimeStart        *time.Time // 解析后的时间
	PushplusToken    string     `env:"PUSHPLUS_TOKEN"`
	Interval         int        `env:"TICKET_INTERVAL" envDefault:"300"`
	GTBaseURL        string     `env:"GT_BASE_URL"`
}

func LoadConfig() *Config {
	cfg := &Config{}
	if err := env.Parse(cfg); err != nil {
		log.Fatalf("环境变量解析失败: %v", err)
	}
	if cfg.MasterServerAddr == "" {
		log.Fatalf("❌ MASTER_SERVER_ADDR 是必需的环境变量，当前未设置")
	}
	if cfg.GTBaseURL == "" {
		log.Fatalf("❌ GT_BASE_URL 是必需的环境变量，当前未设置")
	}
	if cfg.PushplusToken == "" {
		log.Println("⚠️ 未设置 PUSHPLUS_TOKEN，将不会发送推送提醒")
	}
	if cfg.TimeStartRaw == "" {
		log.Println("⚠️ 未设置 TICKET_TIME_START，将不会使用定时抢票")
	} else {
		//设置时间
		loc, _ := time.LoadLocation("Asia/Shanghai")
		TimeStart, err := time.ParseInLocation("2006-01-02T15:04", cfg.TimeStartRaw, loc)
		if err != nil {
			_ = fmt.Errorf("时间格式错误: %v，正确格式应为 2006-01-02T15:04（北京时间）", err)
		}
		cfg.TimeStart = &TimeStart
	}
	if cfg.Interval <= 0 {
		log.Println("⚠️ TICKET_INTERVAL 格式错误（非正数），使用默认值 300")
		cfg.Interval = 300
	} else {
		log.Printf("ℹ️ 抢票重试间隔: %d 秒", cfg.Interval)
	}
	return cfg
}

var Cfg = LoadConfig()
