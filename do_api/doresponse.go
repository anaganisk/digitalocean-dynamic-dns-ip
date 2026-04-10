package do_api

// DOResponse : DigitalOcean DNS Records response.
type DOResponse struct {
	DomainRecords []DNSRecord `json:"domain_records"`
	Meta          struct {
		Total int `json:"total"`
	} `json:"meta"`
	Links struct {
		Pages struct {
			First    string `json:"first"`
			Previous string `json:"prev"`
			Next     string `json:"next"`
			Last     string `json:"last"`
		} `json:"pages"`
	} `json:"links"`
}
