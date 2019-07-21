package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"strconv"

	homedir "github.com/mitchellh/go-homedir"
)

func checkError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

var config ClientConfig

// ClientConfig : configuration json
type ClientConfig struct {
	APIKey     string   `json:"apiKey"`
	DOPageSize int      `json:"doPageSize"`
	UseIPv4    *bool    `json:"useIPv4"`
	UseIPv6    *bool    `json:"useIPv6"`
	Domains    []Domain `json:"domains"`
}

// Domain : domains to be changed
type Domain struct {
	Domain  string      `json:"domain"`
	Records []DNSRecord `json:"records"`
}

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

//GetConfig : get configuration file ~/.digitalocean-dynamic-ip.json
func GetConfig() ClientConfig {
	cmdHelp := flag.Bool("h", false, "Show the help message")
	cmdHelp2 := flag.Bool("help", false, "Show the help message")
	flag.Parse()

	if *cmdHelp || *cmdHelp2 {
		usage()
		os.Exit(1)
	}

	configFile := ""
	if len(flag.Args()) == 0 {
		var err error
		configFile, err = homedir.Dir()
		checkError(err)
		configFile += "/.digitalocean-dynamic-ip.json"
	} else {
		configFile = flag.Args()[0]
	}

	getfile, err := ioutil.ReadFile(configFile)
	checkError(err)
	var config ClientConfig
	json.Unmarshal(getfile, &config)
	checkError(err)
	return config
}

func usage() {
	os.Stdout.WriteString(fmt.Sprintf("To use this program you can specify the following command options:\n"+
		"-h | -help\n\tShow this help message\n"+
		"[config_file]\n\tlocation of the configuration file\n\n"+
		"If the [config_file] parameter is not passed, then the default\n"+
		"config location of ~/.digitalocean-dynamic-ip.json will be used.\n\n"+
		"example usages:\n\t%[1]s -help\n"+
		"\t%[1]s\n"+
		"\t%[1]s %[2]s\n"+
		"",
		os.Args[0],
		"/path/to/my/config.json",
	))
}

//CheckLocalIPs : get current IP of server. checks both IPv4 and Ipv6 to support dual stack environments
func CheckLocalIPs() (ipv4, ipv6 net.IP) {
	var ipv4String, ipv6String string

	if config.UseIPv4 == nil || *(config.UseIPv4) {
		ipv4String, _ = getURLBody("https://ipv4bot.whatismyipaddress.com")
		if ipv4String == "" {
			log.Println("No IPv4 address found. Consider disabling IPv4 checks in the config `\"useIPv4\": false`")
		} else {
			ipv4 = net.ParseIP(ipv4String)
			if ipv4 != nil {
				// make sure we got back an actual ipv4 address
				ipv4 = ipv4.To4()
			}
			if ipv4 == nil {
				log.Printf("Unable to parse `%s` as an IPv4 address\n", ipv4String)
			}
		}
	}

	if config.UseIPv6 == nil || *(config.UseIPv6) {
		ipv6String, _ = getURLBody("https://ipv6bot.whatismyipaddress.com")
		if ipv6String == "" {
			log.Println("No IPv6 address found. Consider disabling IPv6 checks in the config `\"useIPv6\": false`")
		} else {
			ipv6 = net.ParseIP(ipv6String)
			if ipv6 == nil {
				log.Printf("Unable to parse `%s` as an IPv4 address\n", ipv6String)
			}
		}
	}
	return ipv4, ipv6
}

func getURLBody(url string) (string, error) {
	request, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer request.Body.Close()
	body, err := ioutil.ReadAll(request.Body)
	checkError(err)
	return string(body), nil
}

//GetDomainRecords : Get DNS records of current domain.
func GetDomainRecords(domain string) []DNSRecord {
	ret := make([]DNSRecord, 0)
	var page DOResponse
	pageParam := ""
	// 20 is the default page size
	if config.DOPageSize > 0 && config.DOPageSize != 20 {
		pageParam = "?per_page=" + strconv.Itoa(config.DOPageSize)
	}
	for url := "https://api.digitalocean.com/v2/domains/" + url.PathEscape(domain) + "/records" + pageParam; url != ""; url = page.Links.Pages.Next {
		page = getPage(url)
		ret = append(ret, page.DomainRecords...)
	}
	return ret
}

func getPage(url string) DOResponse {
	log.Println(url)
	client := &http.Client{}
	request, err := http.NewRequest("GET", url, nil)
	checkError(err)
	request.Header.Add("Content-type", "Application/json")
	request.Header.Add("Authorization", "Bearer "+config.APIKey)
	response, err := client.Do(request)
	checkError(err)
	defer response.Body.Close()
	body, err := ioutil.ReadAll(response.Body)
	// log.Print(string(body))
	var jsonDOResponse DOResponse
	e := json.Unmarshal(body, &jsonDOResponse)
	checkError(e)
	return jsonDOResponse
}

// UpdateRecords : Update DNS records of domain
func UpdateRecords(domain string, ipv4, ipv6 net.IP, toUpdateRecords []DNSRecord) {
	log.Printf("%s: %d to update\n", domain, len(toUpdateRecords))
	updated := 0
	doRecords := GetDomainRecords(domain)
	// look for the item to update
	if len(doRecords) < 1 {
		log.Printf("%s: No DNS records found", domain)
		return
	}
	log.Printf("%s: %d DNS records found", domain, len(doRecords))
	for _, toUpdateRecord := range toUpdateRecords {
		if toUpdateRecord.Type != "A" && toUpdateRecord.Type != "AAAA" {
			log.Printf("%s: Unsupported type (Only A and AAAA records supported) for updates %+v", domain, toUpdateRecord)
			continue
		}
		if ipv4 == nil && toUpdateRecord.Type == "A" {
			log.Printf("%s: You are trying to update an IPv4 A record with no IPv4 address: config: %+v", domain, toUpdateRecord)
			continue
		}
		if ipv6 == nil && toUpdateRecord.Type == "AAAA" {
			log.Printf("%s: You are trying to update an IPv6 AAAA record with no IPv6 address: config: %+v", domain, toUpdateRecord)
			continue
		}
		if toUpdateRecord.ID > 0 {
			// update the record directly. skip the extra search
			log.Printf("%s: Unable to directly update records yet. Record: %+v", domain, toUpdateRecord)
			continue
		}

		var currentIP string
		if toUpdateRecord.Type == "A" {
			currentIP = ipv4.String()
		} else {
			currentIP = ipv6.String()
		}

		log.Printf("%s: trying to update `%s` : `%s`", domain, toUpdateRecord.Type, toUpdateRecord.Name)
		for _, doRecord := range doRecords {
			//log.Printf("%s: checking `%s` : `%s`", domain, doRecord.Type, doRecord.Name)
			if doRecord.Name == toUpdateRecord.Name && doRecord.Type == toUpdateRecord.Type {
				if doRecord.Data == currentIP && (toUpdateRecord.TTL < 30 || doRecord.TTL == toUpdateRecord.TTL) {
					log.Printf("%s: IP/TTL did not change %+v", domain, doRecord)
					continue
				}
				log.Printf("%s: updating %+v", domain, doRecord)
				// set the IP address
				doRecord.Data = currentIP
				if toUpdateRecord.TTL >= 30 && doRecord.TTL != toUpdateRecord.TTL {
					doRecord.TTL = toUpdateRecord.TTL
				}
				update, err := json.Marshal(doRecord)
				checkError(err)
				client := &http.Client{}
				request, err := http.NewRequest("PUT",
					"https://api.digitalocean.com/v2/domains/"+url.PathEscape(domain)+"/records/"+strconv.FormatInt(int64(doRecord.ID), 10),
					bytes.NewBuffer(update))
				checkError(err)
				request.Header.Set("Content-Type", "application/json")
				request.Header.Add("Authorization", "Bearer "+config.APIKey)
				response, err := client.Do(request)
				checkError(err)
				defer response.Body.Close()
				body, err := ioutil.ReadAll(response.Body)
				log.Printf("%s: DO update response for %s: %s\n", domain, doRecord.Name, string(body))
				updated++
			}
		}

	}
	log.Printf("%s: %d of %d records updated\n", domain, updated, len(toUpdateRecords))
}

func main() {
	config = GetConfig()
	currentIPv4, currentIPv6 := CheckLocalIPs()
	if currentIPv4 == nil && currentIPv6 == nil {
		log.Fatalf("current IP addresses are not a valid, or both are disabled in the config. Check you configuration and internet connection",
			currentIPv4,
			currentIPv6,
		)
	}
	for _, domain := range config.Domains {
		domainName := domain.Domain
		log.Printf("%s: START\n", domainName)
		UpdateRecords(domainName, currentIPv4, currentIPv6, domain.Records)
		log.Printf("%s: END\n", domainName)
	}
}
