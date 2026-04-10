package config

import (
	"encoding/json"
	"os"
	"time"

	"github.com/anaganisk/digitalocean-dynamic-dns-ip/constants"
	"github.com/anaganisk/digitalocean-dynamic-dns-ip/logger"
)

var active ClientConfig
var configFile string

// ClientConfig : configuration json
type ClientConfig struct {
	APIKey                 string   `json:"apiKey"`
	DOPageSize             int      `json:"doPageSize"`
	UseIPv4                *bool    `json:"useIPv4"`
	UseIPv6                *bool    `json:"useIPv6"`
	IPv4CheckURL           string   `json:"ipv4CheckUrl"`
	IPv6CheckURL           string   `json:"ipv6CheckUrl"`
	AllowIPv4InIPv6        bool     `json:"allowIPv4InIPv6"`
	IPvCheckTimeoutSeconds int      `json:"ipvCheckTimeoutSeconds"`
	Domains                []Domain `json:"domains"`
}

// BoundedPageSize returns the configured page size clamped to the valid bounds
func (c ClientConfig) BoundedPageSize() int {
	if c.DOPageSize > 0 && c.DOPageSize != constants.DefaultPageSize {
		pageSize := c.DOPageSize
		// don't let users set more than the max size
		if pageSize > constants.MaxPageSize {
			pageSize = constants.MaxPageSize
		}
		return pageSize
	}
	return constants.DefaultPageSize
}

// getHTTPTimeout returns the configured HTTP timeout or the default if not configured
func (c ClientConfig) GetHTTPTimeout() time.Duration {
	if c.IPvCheckTimeoutSeconds > 0 {
		return time.Duration(c.IPvCheckTimeoutSeconds) * time.Second
	}
	return constants.DefaultIPCheckTimeout
}

// Get loads the configuration from the specified JSON file or default ~/.digitalocean-dynamic-ip.json
func Get() ClientConfig {
	if active.APIKey != "" {
		return active
	}
	return Set(getConfig())
}

func Set(config ClientConfig) ClientConfig {
	active = config
	return active
}

func SetConfigFilePath(path string) {
	configFile = path
}

func getConfig() ClientConfig {
	if configFile == "" {
		logger.Warning("No config file path presented. Using active config, which may be invalid")
		return active
	}

	configData, err := os.ReadFile(configFile)
	logger.CheckError(err)
	var config ClientConfig
	err = json.Unmarshal(configData, &config)
	logger.CheckError(err)
	return config
}
