package do_api

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"net/url"
	"strconv"

	"github.com/anaganisk/digitalocean-dynamic-dns-ip/config"
	"github.com/anaganisk/digitalocean-dynamic-dns-ip/constants"
	"github.com/anaganisk/digitalocean-dynamic-dns-ip/logger"
)

type DNSRecord = config.DNSRecord

// vars pulled out so they can be overriden in tests
var digitalOceanAPIBase = "https://api.digitalocean.com/v2"

// GetDomainRecords : Get DNS records of current domain.
// Uses the DigitalOcean Domains API list records endpoint:
// https://docs.digitalocean.com/reference/api/reference/domain-records/#domains_list_records
func GetDomainRecords(domain config.Domain) []config.DNSRecord {
	records := make([]DNSRecord, 0)
	var page DOResponse
	pageParam := ""
	// Use configured page size if available and different from default
	if pageSize := config.Get().BoundedPageSize(); pageSize != constants.DefaultPageSize {
		pageParam = "?per_page=" + strconv.Itoa(pageSize)
	}
	// Check if all records have the same type and add type filter if so
	if recordType := domain.UniformRecordType(); recordType != "" {
		separator := "?"
		if pageParam != "" {
			separator = "&"
		}
		pageParam += separator + "type=" + url.QueryEscape(recordType)
	}
	for url := digitalOceanAPIBase + "/domains/" + url.PathEscape(domain.Domain) + "/records" + pageParam; url != ""; url = page.Links.Pages.Next {
		page = getPage(url)
		records = append(records, page.DomainRecords...)
	}
	return records
}

// getPage fetches a page of DNS records from DigitalOcean API
func getPage(url string) DOResponse {
	logger.Debug("%s", url)
	client := &http.Client{Timeout: config.Get().GetHTTPTimeout()}
	request, err := http.NewRequest("GET", url, nil)
	logger.CheckError(err)
	request.Header.Add("Content-type", "Application/json")
	request.Header.Add("Authorization", "Bearer "+config.Get().APIKey)
	response, err := client.Do(request)
	logger.CheckError(err)
	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)
	logger.CheckError(err)
	// logger.Debug("%s", string(body))
	var jsonDOResponse DOResponse
	err = json.Unmarshal(body, &jsonDOResponse)
	logger.CheckError(err)
	return jsonDOResponse
}

// UpdateRecords : Update DNS records of domain.
// Uses the DigitalOcean Domains API update record endpoint:
// https://docs.digitalocean.com/reference/api/reference/domain-records/#domains_update_record
func UpdateRecords(domain config.Domain, ipv4, ipv6 net.IP) {
	logger.Debug("%s: %d to update", domain.Domain, len(domain.Records))
	updated := 0
	doRecords := GetDomainRecords(domain)
	// look for the item to update
	if len(doRecords) < 1 {
		logger.Warningf("%s: No DNS records found in DigitalOcean", domain.Domain)
		return
	}
	logger.Debug("%s: %d DNS records found in DigitalOcean", domain.Domain, len(doRecords))
	for _, toUpdateRecord := range domain.Records {
		if !toUpdateRecord.IsValidType() {
			logger.Warningf("%s: Unsupported type (Only A and AAAA records supported) for updates %+v", domain.Domain, toUpdateRecord)
			continue
		}
		if ipv4 == nil && toUpdateRecord.Type == "A" {
			logger.Warningf("%s: You are trying to update an IPv4 A record with no IPv4 address: config: %+v", domain.Domain, toUpdateRecord)
			continue
		}
		if toUpdateRecord.ID > 0 {
			// update the record directly. skip the extra search
			logger.Warningf("%s: Unable to directly update records yet. Record: %+v", domain.Domain, toUpdateRecord)
			continue
		}

		var currentIP string
		if toUpdateRecord.Type == "A" {
			currentIP = ipv4.String()
		} else if ipv6 == nil || ipv6.To4() != nil {
			if ipv6 == nil {
				ipv6 = ipv4
			}

			logger.Warningf("%s: You are trying to update an IPv6 AAAA record without an IPv6 address: ip: %s config: %+v",
				domain.Domain,
				ipv6,
				toUpdateRecord,
			)
			if config.Get().AllowIPv4InIPv6 {
				currentIP = toIPv6String(ipv6)
				logger.Debug("%s: Converting IPv4 `%s` to IPv6 `%s`", domain.Domain, ipv6.String(), currentIP)
			} else {
				continue
			}
		} else {
			currentIP = ipv6.String()
		}

		logger.Debug("%s: trying to update `%s` : `%s`", domain.Domain, toUpdateRecord.Type, toUpdateRecord.Name)
		for _, doRecord := range doRecords {
			//logger.Debug("%s: checking `%s` : `%s`", domain.Domain, doRecord.Type, doRecord.Name)
			if doRecord.Name == toUpdateRecord.Name && doRecord.Type == toUpdateRecord.Type {
				if doRecord.Data == currentIP && (toUpdateRecord.TTL < constants.MinTTL || doRecord.TTL == toUpdateRecord.TTL) {
					logger.Debug("%s: IP/TTL did not change %+v", domain.Domain, doRecord)
					continue
				}
				logger.Debug("%s: updating %+v", domain.Domain, doRecord)
				// set the IP address
				doRecord.Data = currentIP
				if toUpdateRecord.IsValidTTL() && doRecord.TTL != toUpdateRecord.TTL {
					doRecord.TTL = toUpdateRecord.TTL
				}
				update, err := json.Marshal(doRecord)
				logger.CheckError(err)
				client := &http.Client{}
				request, err := http.NewRequest("PUT",
					digitalOceanAPIBase+"/domains/"+url.PathEscape(domain.Domain)+"/records/"+strconv.FormatInt(int64(doRecord.ID), 10),
					bytes.NewBuffer(update))
				logger.CheckError(err)
				request.Header.Set("Content-Type", "application/json")
				request.Header.Add("Authorization", "Bearer "+config.Get().APIKey)
				response, err := client.Do(request)
				logger.CheckError(err)
				defer response.Body.Close()
				body, err := io.ReadAll(response.Body)
				logger.CheckError(err)
				logger.Debug("%s: DO update response for %s: %s", domain.Domain, doRecord.Name, string(body))
				updated++
			}
		}

	}
	logger.Debug("%s: %d of %d records updated", domain.Domain, updated, len(domain.Records))
}

// toIPv6String : net.IP.String will always output an IPv4 address in dot
// notation (127.0.0.1) even if we convert it using net.IP.To16().
// For AAAA records, we can't have that. Instead, force the
// IP to have the IPv6 colon notation.
func toIPv6String(ip net.IP) (currentIP string) {
	if ip == nil {
		return ""
	}
	if ipv4 := ip.To4(); ipv4 != nil {
		ip = ipv4
	}
	l := len(ip)
	if l < 16 {
		// ensure "v4InV6Prefix" for IPv4 addresses
		currentIP = "::ffff:"
	}
	// byte length of an ipv6 segment.
	segSize := 2
	for i := 0; i < l; i += segSize {
		end := i + segSize
		bs := ip[i:end]
		addColon := (end + 1) < l
		currentIP += hex.EncodeToString(bs)
		if addColon {
			currentIP += ":"
		}
	}
	return currentIP
}
