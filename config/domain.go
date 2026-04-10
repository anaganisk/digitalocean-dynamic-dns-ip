package config

// Domain : domains to be changed
type Domain struct {
	Domain  string      `json:"domain"`
	Records []DNSRecord `json:"records"`
}

// HasUniformRecordType checks if all records in the domain have the same DNS record type
func (d Domain) HasUniformRecordType() bool {
	if len(d.Records) == 0 {
		return false
	}
	recordType := d.Records[0].Type
	for _, record := range d.Records {
		if record.Type != recordType {
			return false
		}
	}
	return true
}

// UniformRecordType returns the record type if all records have the same type, empty string otherwise
func (d Domain) UniformRecordType() string {
	if !d.HasUniformRecordType() {
		return ""
	}
	return d.Records[0].Type
}
