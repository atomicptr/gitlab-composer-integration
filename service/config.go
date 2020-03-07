package service

import (
	"net/url"
	"strings"
	"time"

	"github.com/pkg/errors"
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
	HttpCredentials     string        `conf:""`
}

// Validate the configuration
func (config *Config) Validate() error {
	_, err := url.Parse(config.GitlabUrl)
	if err != nil {
		return err
	}

	if len(config.HttpCredentials) > 0 && !strings.Contains(config.HttpCredentials, ":") {
		return errors.New("http credentials should be in the form of \"username:password\" or empty.")
	}

	return nil
}

// IsVendorAllowed checks if the given vendor is allowed
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

// GetHttpCredentials returns username:password combination as username, password pair
func (config *Config) GetHttpCredentials() (string, string) {
	parts := strings.Split(config.HttpCredentials, ":")
	if len(parts) < 2 {
		return "", ""
	}
	// return everything from the first match and the rest
	return parts[0], config.HttpCredentials[len(parts[0])+1:]
}
