package main

import (
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/anaganisk/digitalocean-dynamic-dns-ip/config"
)

// boolPtr is a helper function to create a pointer to a boolean value
func boolPtr(v bool) *bool {
	return &v
}

func TestCheckLocalIPsIPv4Only(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("192.0.2.1"))
	}))
	defer server.Close()

	oldConfig := config.Get()
	config.Set(config.ClientConfig{
		UseIPv4:                boolPtr(true),
		UseIPv6:                boolPtr(false),
		IPv4CheckURL:           server.URL,
		IPv6CheckURL:           "https://invalid.example",
		IPvCheckTimeoutSeconds: 5,
	})
	defer func() { config.Set(oldConfig) }()

	ipv4, ipv6 := CheckPublicIPs()
	if ipv4 == nil || ipv4.String() != "192.0.2.1" {
		t.Fatalf("CheckLocalIPs() ipv4 = %v, want 192.0.2.1", ipv4)
	}
	if ipv6 != nil {
		t.Fatalf("CheckLocalIPs() ipv6 = %v, want nil", ipv6)
	}
}

// TestGetURLBodyWithMock tests getURLBody with a mocked HTTP server
func TestGetURLBodyWithMock(t *testing.T) {
	// Create a mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("192.168.1.100"))
	}))
	defer server.Close()

	// Temporarily set config for this test
	oldConfig := config.Get()
	config.Set(config.ClientConfig{
		IPvCheckTimeoutSeconds: 5,
	})
	defer func() { config.Set(oldConfig) }()

	result, err := getURLBody(server.URL)
	if err != nil {
		t.Fatalf("getURLBody() error = %v", err)
	}

	expected := "192.168.1.100"
	if strings.TrimSpace(result) != expected {
		t.Errorf("getURLBody() = %s, want %s", strings.TrimSpace(result), expected)
	}
}

// TestParseIPFromResponse tests parsing an IP address from an HTTP response
func TestParseIPFromResponse(t *testing.T) {
	tests := []struct {
		name       string
		response   string
		shouldFail bool
		expected   net.IP
	}{
		{
			name:       "valid IPv4",
			response:   "192.168.1.100",
			shouldFail: false,
			expected:   net.ParseIP("192.168.1.100"),
		},
		{
			name:       "valid IPv6",
			response:   "2001:db8::1",
			shouldFail: false,
			expected:   net.ParseIP("2001:db8::1"),
		},
		{
			name:       "invalid IP",
			response:   "not-an-ip",
			shouldFail: false,
			expected:   nil,
		},
		{
			name:       "empty response",
			response:   "",
			shouldFail: false,
			expected:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := net.ParseIP(strings.TrimSpace(tt.response))
			if (result == nil && tt.expected != nil) || (result != nil && tt.expected == nil) {
				t.Errorf("ParseIP() = %v, want %v", result, tt.expected)
			}
			if result != nil && tt.expected != nil && !result.Equal(tt.expected) {
				t.Errorf("ParseIP() = %v, want %v", result, tt.expected)
			}
		})
	}
}
