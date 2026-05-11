package config

import (
	"errors"
	"log"
	"os"
)

// Config holds all application configuration loaded from environment variables.
type Config struct {
	MYSQDSN          string
	JWTSecret        string
	JWTRefreshSecret string
	WXAppID          string
	WXSecret         string
	Port             string
	FeishuAppID      string
	FeishuAppSecret  string
	DeepSeekAPIKey   string
}

// Load reads environment variables and returns a populated Config.
// Required vars: MYSQL_DSN, JWT_SECRET, JWT_REFRESH_SECRET.
func Load() (*Config, error) {
	cfg := &Config{}

	cfg.MYSQDSN = os.Getenv("MYSQL_DSN")
	if cfg.MYSQDSN == "" {
		return nil, errors.New("required env var MYSQL_DSN is not set")
	}

	cfg.JWTSecret = os.Getenv("JWT_SECRET")
	if cfg.JWTSecret == "" {
		return nil, errors.New("required env var JWT_SECRET is not set")
	}

	cfg.JWTRefreshSecret = os.Getenv("JWT_REFRESH_SECRET")
	if cfg.JWTRefreshSecret == "" {
		return nil, errors.New("required env var JWT_REFRESH_SECRET is not set")
	}

	cfg.WXAppID = os.Getenv("WX_APPID")
	if cfg.WXAppID == "" {
		log.Println("warning: optional env var WX_APPID is not set")
	}

	cfg.WXSecret = os.Getenv("WX_SECRET")
	if cfg.WXSecret == "" {
		log.Println("warning: optional env var WX_SECRET is not set")
	}

	cfg.Port = os.Getenv("PORT")
	if cfg.Port == "" {
		cfg.Port = "8080"
	}

	cfg.FeishuAppID = os.Getenv("FEISHU_APP_ID")
	cfg.FeishuAppSecret = os.Getenv("FEISHU_APP_SECRET")
	cfg.DeepSeekAPIKey = os.Getenv("DEEPSEEK_API_KEY")

	return cfg, nil
}
