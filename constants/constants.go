package constants

import "time"

const (
	DefaultIPCheckTimeout = 15 * time.Second // Default timeout for IP check requests
	MinTTL                = 30               // Minimum TTL allowed by DigitalOcean
	DefaultPageSize       = 20               // Default page size for DigitalOcean API
	MaxPageSize           = 200              // Maximum page size for DigitalOcean API

	DefaultIPv4CheckURL = "https://api.ipify.org/?format=text"
	DefaultIPv6CheckURL = "https://api64.ipify.org/?format=text"
)
