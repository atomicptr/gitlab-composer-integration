package service

import "time"

type Config struct {
	Port        int           `conf:"default:4000"`
	HttpTimeout time.Duration `conf:"default:30s"`
}
