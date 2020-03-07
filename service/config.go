package service

import (
	"net/url"
	"time"
)

type Config struct {
	GitlabUrl           string        `conf:"required"`
	GitlabToken         string        `conf:"required,noprint"`
	CacheExpireDuration time.Duration `conf:"default:60m"`
	CacheFilePath       string        `conf:""`
	VendorWhitelist     []string      `conf:""`
	Port                int           `conf:"default:4000"`
	HttpTimeout         time.Duration `conf:"default:30s"`
	NoCache             bool          `conf:"default:false"`
}

func (config *Config) Validate() error {
	_, err := url.Parse(config.GitlabUrl)
	if err != nil {
		return err
	}

	return nil
}

func (config *Config) IsVendorAllowed(vendorName string) bool {
	// vendor whitelist is empty, allow everything
	if len(config.VendorWhitelist) == 0 {
		return true
	}

	// vendor whitelist is enabled, only allow specified vendors
	for _, name := range config.VendorWhitelist {
		if name == vendorName {
			return true
		}
	}

	return false
}
