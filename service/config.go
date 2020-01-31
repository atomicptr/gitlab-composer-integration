package service

import (
	"net/url"
	"time"
)

type Config struct {
	GitlabUrl           string        `conf:"required"`
	GitlabToken         string        `conf:"required,noprint"`
	CacheExpireDuration time.Duration `conf:"default:60m"`
	Port                int           `conf:"default:4000"`
	HttpTimeout         time.Duration `conf:"default:30s"`
}

func (config *Config) Validate() error {
	_, err := url.Parse(config.GitlabUrl)
	if err != nil {
		return err
	}

	return nil
}
