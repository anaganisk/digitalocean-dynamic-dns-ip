package config

import (
	"github.com/anaganisk/digitalocean-dynamic-dns-ip/constants"
)

// DNSRecord : Modifyiable DNS record
type DNSRecord struct {
	ID       int64   `json:"id"`
	Type     string  `json:"type"`
	Name     string  `json:"name"`
	Priority *int    `json:"priority"`
	Port     *int    `json:"port"`
	Weight   *int    `json:"weight"`
	TTL      int     `json:"ttl"`
	Flags    *uint8  `json:"flags"`
	Tag      *string `json:"tag"`
	Data     string  `json:"data"`
}

// IsValidType checks if the DNS record type is supported (A or AAAA)
func (r DNSRecord) IsValidType() bool {
	return r.Type == "A" || r.Type == "AAAA"
}

// IsValidTTL checks if the DNS record TTL value meets the minimum requirement
func (r DNSRecord) IsValidTTL() bool {
	return r.TTL >= constants.MinTTL
}
