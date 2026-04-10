package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/anaganisk/digitalocean-dynamic-dns-ip/constants"
)

func TestGetConfigFromFile(t *testing.T) {
	configJSON := `{
		"apiKey": "testkey",
		"doPageSize": 40,
		"useIPv4": true,
		"useIPv6": false,
		"ipv4CheckUrl": "https://example.com/ipv4",
		"ipv6CheckUrl": "https://example.com/ipv6",
		"allowIPv4InIPv6": false,
		"domains": []
	}`

	configPath := filepath.Join(t.TempDir(), "config.json")
	if err := os.WriteFile(configPath, []byte(configJSON), 0600); err != nil {
		t.Fatalf("unable to write config file: %v", err)
	}

	SetConfigFilePath(configPath)

	cfg := Get()
	if cfg.APIKey != "testkey" {
		t.Fatalf("GetConfig() APIKey = %s, want testkey", cfg.APIKey)
	}
	if cfg.DOPageSize != 40 {
		t.Fatalf("GetConfig() DOPageSize = %d, want 40", cfg.DOPageSize)
	}
}

// TestClientConfigPageSizeBounding tests the BoundedPageSize receiver method
func TestClientConfigPageSizeBounding(t *testing.T) {
	tests := []struct {
		name     string
		config   ClientConfig
		expected int
	}{
		{
			name:     "default page size",
			config:   ClientConfig{DOPageSize: 0},
			expected: constants.DefaultPageSize,
		},
		{
			name:     "custom valid page size",
			config:   ClientConfig{DOPageSize: 50},
			expected: 50,
		},
		{
			name:     "page size exceeds max",
			config:   ClientConfig{DOPageSize: 300},
			expected: constants.MaxPageSize,
		},
		{
			name:     "page size equals default",
			config:   ClientConfig{DOPageSize: constants.DefaultPageSize},
			expected: constants.DefaultPageSize,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.config.BoundedPageSize()
			if result != tt.expected {
				t.Errorf("BoundedPageSize() = %d, want %d", result, tt.expected)
			}
		})
	}
}

// TestDomainHasUniformRecordType tests the HasUniformRecordType receiver method
func TestDomainHasUniformRecordType(t *testing.T) {
	tests := []struct {
		name     string
		domain   Domain
		expected bool
	}{
		{
			name:     "empty domain",
			domain:   Domain{Domain: "test", Records: []DNSRecord{}},
			expected: false,
		},
		{
			name: "single A record",
			domain: Domain{
				Domain:  "test",
				Records: []DNSRecord{{Type: "A"}},
			},
			expected: true,
		},
		{
			name: "multiple A records",
			domain: Domain{
				Domain: "test",
				Records: []DNSRecord{
					{Type: "A", Name: "www"},
					{Type: "A", Name: "api"},
				},
			},
			expected: true,
		},
		{
			name: "mixed A and AAAA records",
			domain: Domain{
				Domain: "test",
				Records: []DNSRecord{
					{Type: "A", Name: "www"},
					{Type: "AAAA", Name: "www"},
				},
			},
			expected: false,
		},
		{
			name: "multiple AAAA records",
			domain: Domain{
				Domain: "test",
				Records: []DNSRecord{
					{Type: "AAAA", Name: "www"},
					{Type: "AAAA", Name: "api"},
				},
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.domain.HasUniformRecordType()
			if result != tt.expected {
				t.Errorf("HasUniformRecordType() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// TestDomainUniformRecordType tests the UniformRecordType receiver method
func TestDomainUniformRecordType(t *testing.T) {
	tests := []struct {
		name     string
		domain   Domain
		expected string
	}{
		{
			name:     "empty domain",
			domain:   Domain{Domain: "test", Records: []DNSRecord{}},
			expected: "",
		},
		{
			name: "single A record",
			domain: Domain{
				Domain:  "test",
				Records: []DNSRecord{{Type: "A"}},
			},
			expected: "A",
		},
		{
			name: "multiple A records",
			domain: Domain{
				Domain: "test",
				Records: []DNSRecord{
					{Type: "A", Name: "www"},
					{Type: "A", Name: "api"},
				},
			},
			expected: "A",
		},
		{
			name: "mixed A and AAAA records",
			domain: Domain{
				Domain: "test",
				Records: []DNSRecord{
					{Type: "A", Name: "www"},
					{Type: "AAAA", Name: "www"},
				},
			},
			expected: "",
		},
		{
			name: "multiple AAAA records",
			domain: Domain{
				Domain: "test",
				Records: []DNSRecord{
					{Type: "AAAA", Name: "www"},
					{Type: "AAAA", Name: "api"},
				},
			},
			expected: "AAAA",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.domain.UniformRecordType()
			if result != tt.expected {
				t.Errorf("UniformRecordType() = %s, want %s", result, tt.expected)
			}
		})
	}
}

// TestDNSRecordTypeValidation tests the IsValidType receiver method
func TestDNSRecordTypeValidation(t *testing.T) {
	tests := []struct {
		name    string
		record  DNSRecord
		isValid bool
	}{
		{
			name:    "A record",
			record:  DNSRecord{Type: "A"},
			isValid: true,
		},
		{
			name:    "AAAA record",
			record:  DNSRecord{Type: "AAAA"},
			isValid: true,
		},
		{
			name:    "MX record (unsupported)",
			record:  DNSRecord{Type: "MX"},
			isValid: false,
		},
		{
			name:    "CNAME record (unsupported)",
			record:  DNSRecord{Type: "CNAME"},
			isValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.record.IsValidType()
			if result != tt.isValid {
				t.Errorf("IsValidType() = %v, want %v", result, tt.isValid)
			}
		})
	}
}

// TestDNSRecordTTLValidation tests the IsValidTTL receiver method
func TestDNSRecordTTLValidation(t *testing.T) {
	tests := []struct {
		name    string
		record  DNSRecord
		isValid bool
	}{
		{
			name:    "minimum TTL",
			record:  DNSRecord{TTL: 30},
			isValid: true,
		},
		{
			name:    "TTL below minimum",
			record:  DNSRecord{TTL: 15},
			isValid: false,
		},
		{
			name:    "typical TTL",
			record:  DNSRecord{TTL: 3600},
			isValid: true,
		},
		{
			name:    "high TTL",
			record:  DNSRecord{TTL: 86400},
			isValid: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.record.IsValidTTL()
			if result != tt.isValid {
				t.Errorf("IsValidTTL() = %v, want %v", result, tt.isValid)
			}
		})
	}
}

// TestHTTPTimeout verifies that getHTTPTimeout returns the appropriate timeout value
func TestHTTPTimeout(t *testing.T) {
	tests := []struct {
		name       string
		configTime int
		expected   time.Duration
	}{
		{
			name:       "custom timeout",
			configTime: 30,
			expected:   30 * time.Second,
		},
		{
			name:       "zero timeout uses default",
			configTime: 0,
			expected:   15 * time.Second,
		},
		{
			name:       "negative timeout uses default",
			configTime: -5,
			expected:   15 * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conf := ClientConfig{
				IPvCheckTimeoutSeconds: tt.configTime,
			}

			result := conf.GetHTTPTimeout()
			if result != tt.expected {
				t.Errorf("getHTTPTimeout() = %v, want %v", result, tt.expected)
			}
		})
	}
}
