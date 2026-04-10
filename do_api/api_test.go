package do_api

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/anaganisk/digitalocean-dynamic-dns-ip/config"
)

func TestGetDomainRecordsPagination(t *testing.T) {
	var server *httptest.Server
	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/domains/test/records" {
			t.Fatalf("unexpected path %s", r.URL.Path)
		}

		query := r.URL.RawQuery

		if query == "per_page=2&type=A" {
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(fmt.Sprintf(`{
				"domain_records": [
					{"id": 1, "type": "A", "name": "www", "data": "1.1.1.1", "ttl": 3600, "priority": null, "port": null, "weight": null, "flags": null, "tag": null}
				],
				"meta": {"total": 2},
				"links": {"pages": {"next": "%s/domains/test/records?page=2&type=A", "prev": "", "first": "", "last": ""}}
			}`, server.URL)))
			return
		}

		if query == "page=2&type=A" {
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{
				"domain_records": [
					{"id": 2, "type": "A", "name": "api", "data": "1.1.1.2", "ttl": 3600, "priority": null, "port": null, "weight": null, "flags": null, "tag": null}
				],
				"meta": {"total": 2},
				"links": {"pages": {"next": "", "prev": "", "first": "", "last": ""}}
			}`))
			return
		}

		t.Fatalf("unexpected query %s", query)
	}))
	defer server.Close()

	oldBase := digitalOceanAPIBase
	oldConfig := config.Get()
	digitalOceanAPIBase = server.URL
	config.Set(config.ClientConfig{DOPageSize: 2, IPvCheckTimeoutSeconds: 5})
	defer func() {
		digitalOceanAPIBase = oldBase
		config.Set(oldConfig)
	}()

	records := GetDomainRecords(config.Domain{
		Domain: "test",
		Records: []config.DNSRecord{
			{Type: "A", Name: "www"},
			{Type: "A", Name: "api"},
		},
	})
	if len(records) != 2 {
		t.Fatalf("GetDomainRecords() returned %d records, want 2", len(records))
	}
	if records[0].Name != "www" || records[1].Name != "api" {
		t.Fatalf("unexpected records %+v", records)
	}
}

func TestUpdateRecordsPutsRecord(t *testing.T) {
	var updateBody []byte
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/domains/test/records":
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{
				"domain_records": [
					{"id": 1, "type": "A", "name": "www", "data": "1.1.1.1", "ttl": 3600, "priority": null, "port": null, "weight": null, "flags": null, "tag": null}
				],
				"meta": {"total": 1},
				"links": {"pages": {"next": "", "prev": "", "first": "", "last": ""}}
			}`))
		case r.Method == http.MethodPut && r.URL.Path == "/domains/test/records/1":
			var err error
			updateBody, err = io.ReadAll(r.Body)
			if err != nil {
				t.Fatalf("unable to read request body: %v", err)
			}
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"domain_record": {"id": 1}}`))
		default:
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.Path)
		}
	}))
	defer server.Close()

	oldBase := digitalOceanAPIBase
	oldConfig := config.Get()
	digitalOceanAPIBase = server.URL
	config.Set(config.ClientConfig{APIKey: "test-key", IPvCheckTimeoutSeconds: 5})
	defer func() {
		digitalOceanAPIBase = oldBase
		config.Set(oldConfig)
	}()

	UpdateRecords(config.Domain{
		Domain:  "test",
		Records: []config.DNSRecord{{Type: "A", Name: "www", TTL: 300}},
	}, net.ParseIP("1.2.3.4"), nil)

	if updateBody == nil {
		t.Fatal("expected update request body")
	}
	if !strings.Contains(string(updateBody), `"data":"1.2.3.4"`) {
		t.Fatalf("unexpected update body %s", updateBody)
	}
	if !strings.Contains(string(updateBody), `"ttl":300`) {
		t.Fatalf("unexpected update body %s", updateBody)
	}
}

// TestGetPageWithMock tests getPage with a mocked HTTP response
func TestGetPageWithMock(t *testing.T) {
	// Create a mock server that responds with DigitalOcean DNS API format
	responseBody := `{
		"domain_records": [
			{
				"id": 1,
				"type": "A",
				"name": "example",
				"data": "192.168.1.1",
				"priority": null,
				"port": null,
				"ttl": 3600,
				"weight": null,
				"flags": null,
				"tag": null
			}
		],
		"meta": {
			"total": 1
		},
		"links": {
			"pages": {
				"first": "",
				"prev": "",
				"next": "",
				"last": ""
			}
		}
	}`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		io.Copy(w, bytes.NewReader([]byte(responseBody)))
	}))
	defer server.Close()

	// Set config for test
	oldConfig := config.Get()
	config.Set(config.ClientConfig{
		APIKey:                 "test-key",
		IPvCheckTimeoutSeconds: 5,
	})
	defer func() { config.Set(oldConfig) }()

	result := getPage(server.URL)

	if len(result.DomainRecords) != 1 {
		t.Errorf("getPage() returned %d records, want 1", len(result.DomainRecords))
	}

	if result.DomainRecords[0].Name != "example" {
		t.Errorf("Record name = %s, want 'example'", result.DomainRecords[0].Name)
	}

	if result.DomainRecords[0].Type != "A" {
		t.Errorf("Record type = %s, want 'A'", result.DomainRecords[0].Type)
	}
}

// TestToIPv6String tests the IPv6 string conversion function
func TestToIPv6String(t *testing.T) {
	tests := []struct {
		name     string
		input    net.IP
		expected string
	}{
		{
			name:     "nil IP",
			input:    nil,
			expected: "",
		},
		{
			name:     "IPv4 address",
			input:    net.ParseIP("192.168.1.1"),
			expected: "::ffff:c0a8:0101",
		},
		{
			name:     "IPv6 address",
			input:    net.ParseIP("2001:db8::1"),
			expected: "2001:0db8:0000:0000:0000:0000:0000:0001",
		},
		{
			name:     "IPv6 loopback",
			input:    net.ParseIP("::1"),
			expected: "0000:0000:0000:0000:0000:0000:0000:0001",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := toIPv6String(tt.input)
			if result != tt.expected {
				t.Errorf("toIPv6String(%v) = %s, want %s", tt.input, result, tt.expected)
			}
		})
	}
}
