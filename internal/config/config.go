// Package config 负责从环境变量加载服务配置。
package config

import (
	"fmt"
	"os"
	"strconv"
)

// Config 服务配置。
type Config struct {
	Addr         string
	MaxPageSize  int
	APIKey       string
	RateLimit    int
	RateWindowSec int
}

// Load 从环境变量加载配置。
func Load() *Config {
	cfg := &Config{
		Addr:          ":" + getenv("PORT", "8080"),
		MaxPageSize:   getenvInt("MAX_PAGE_SIZE", 100),
		APIKey:        getenv("API_KEY", "loadtest-secret-key"),
		RateLimit:     getenvInt("RATE_LIMIT", 1000),
		RateWindowSec: getenvInt("RATE_WINDOW_SEC", 60),
	}
	if v := os.Getenv("ADDR"); v != "" {
		cfg.Addr = v
	}
	return cfg
}

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func getenvInt(key string, def int) int {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	n, err := strconv.Atoi(v)
	if err != nil || n <= 0 {
		return def
	}
	return n
}

// String 返回配置摘要。
func (c *Config) String() string {
	return fmt.Sprintf("addr=%s max_page_size=%d rate_limit=%d rate_window_sec=%d",
		c.Addr, c.MaxPageSize, c.RateLimit, c.RateWindowSec)
}
