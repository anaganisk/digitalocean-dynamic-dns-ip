package main

import (
	"bytes"
	"encoding/hex"
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
		log.SetOutput(os.Stderr)
		log.Fatal(err)
	}
}

// logError Logs an error message to Stderr and exits the program
func logError(msg string) {
	log.SetOutput(os.Stderr)
	log.Fatal(msg)
}

// logWarning Logs a warning message to Stderr without exiting the program
func logWarning(msg string) {
	old := log.Default().Writer()
	log.SetOutput(os.Stderr)
	log.Println(msg)
	log.SetOutput(old)
}

// logWarning Logs a warning message to Stderr without exiting the program
func logWarningf(format string, v ...interface{}) {
	old := log.Default().Writer()
	log.SetOutput(os.Stderr)
	log.Printf(format, v...)
	log.SetOutput(old)
}

var config ClientConfig

// ClientConfig : configuration json
type ClientConfig struct {
	APIKey          string   `json:"apiKey"`
	DOPageSize      int      `json:"doPageSize"`
	UseIPv4         *bool    `json:"useIPv4"`
	UseIPv6         *bool    `json:"useIPv6"`
	IPv4CheckURL    string   `json:"ipv4CheckUrl"`
	IPv6CheckURL    string   `json:"ipv6CheckUrl"`
	AllowIPv4InIPv6 bool     `json:"allowIPv4InIPv6"`
	Domains         []Domain `json:"domains"`
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
	cmdDbg := flag.Bool("d", false, "Outputs log messages to the standard console")
	cmdDbg2 := flag.Bool("debug", false, "Outputs log messages to the standard console")
	flag.Parse()

	if *cmdHelp || *cmdHelp2 {
		usage()
		os.Exit(1)
	}

	if !((*cmdDbg) || (*cmdDbg2)) {
		// if no debug option was selected, discard all debug output
		log.SetOutput(ioutil.Discard)
	} else {
		// default debug output to Stdout instead of Stderr
		log.SetOutput(os.Stdout)
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

	log.Printf("Using Config file: %s", configFile)

	getfile, err := ioutil.ReadFile(configFile)
	checkError(err)
	var config ClientConfig
	err = json.Unmarshal(getfile, &config)
	checkError(err)
	return config
}

func usage() {
	os.Stdout.WriteString(fmt.Sprintf("To use this program you can specify the following command options:\n"+
		"-h | -help\n\tShow this help message\n"+
		"-d | -debug\n\tPrint debug messages to standard output\n"+
		"[config_file]\n\tlocation of the configuration file\n\n"+
		"If the [config_file] parameter is not passed, then the default\n"+
		"config location of ~/.digitalocean-dynamic-ip.json will be used.\n\n"+
		"example usages:\n\t%[1]s -help\n"+
		"\t%[1]s\n"+
		"\t%[1]s %[2]s\n"+
		"\t%[1]s -debug %[2]s\n"+
		"",
		os.Args[0],
		"/path/to/my/config.json",
	))
}

//CheckLocalIPs : get current IP of server. checks both IPv4 and Ipv6 to support dual stack environments
func CheckLocalIPs() (ipv4, ipv6 net.IP) {
	var ipv4String, ipv6String string
	ipv4CheckURL := "https://api.ipify.org/?format=text"
	ipv6CheckURL := "https://api64.ipify.org/?format=text"
	if len(config.IPv4CheckURL) > 0 {
		ipv4CheckURL = config.IPv4CheckURL
	}
	if len(config.IPv6CheckURL) > 0 {
		ipv6CheckURL = config.IPv6CheckURL
	}

	if config.UseIPv4 == nil || *(config.UseIPv4) {
		log.Printf("Checking IPv4 with URL: %s", ipv4CheckURL)
		ipv4String, _ = getURLBody(ipv4CheckURL)
		if ipv4String == "" {
			logWarning("No IPv4 address found. Consider disabling IPv4 checks in the config `\"useIPv4\": false`")
		} else {
			ipv4 = net.ParseIP(ipv4String)
			if ipv4 != nil {
				// make sure we got back an actual ipv4 address
				ipv4 = ipv4.To4()
				log.Printf("Discovered IPv4 address `%s`", ipv4.String())
			}
			if ipv4 == nil {
				logWarningf("Unable to parse `%s` as an IPv4 address", ipv4String)
			}
		}
	}

	if config.UseIPv6 == nil || *(config.UseIPv6) {
		log.Printf("Checking IPv6 with URL: %s", ipv6CheckURL)
		ipv6String, _ = getURLBody(ipv6CheckURL)
		if ipv6String == "" {
			logWarning("No IPv6 address found. Consider disabling IPv6 checks in the config `\"useIPv6\": false`")
		} else {
			ipv6 = net.ParseIP(ipv6String)
			if ipv6 == nil {
				logWarningf("Unable to parse `%s` as an IPv6 address", ipv6String)
			} else {
				log.Printf("Discovered IPv6 address `%s`", ipv6.String())
			}
		}
	}
	return ipv4, ipv6
}

func getURLBody(url string) (string, error) {
	request, err := http.Get(url)
	checkError(err)
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
		pageSize := config.DOPageSize
		// don't let users set more than the max size
		if pageSize > 200 {
			pageSize = 200
		}
		pageParam = "?per_page=" + strconv.Itoa(pageSize)
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
	checkError(err)
	// log.Println(string(body))
	var jsonDOResponse DOResponse
	err = json.Unmarshal(body, &jsonDOResponse)
	checkError(err)
	return jsonDOResponse
}

// UpdateRecords : Update DNS records of domain
func UpdateRecords(domain Domain, ipv4, ipv6 net.IP) {
	log.Printf("%s: %d to update", domain.Domain, len(domain.Records))
	updated := 0
	doRecords := GetDomainRecords(domain.Domain)
	// look for the item to update
	if len(doRecords) < 1 {
		logWarningf("%s: No DNS records found in DigitalOcean", domain.Domain)
		return
	}
	log.Printf("%s: %d DNS records found in DigitalOcean", domain.Domain, len(doRecords))
	for _, toUpdateRecord := range domain.Records {
		if toUpdateRecord.Type != "A" && toUpdateRecord.Type != "AAAA" {
			logWarningf("%s: Unsupported type (Only A and AAAA records supported) for updates %+v", domain.Domain, toUpdateRecord)
			continue
		}
		if ipv4 == nil && toUpdateRecord.Type == "A" {
			logWarningf("%s: You are trying to update an IPv4 A record with no IPv4 address: config: %+v", domain.Domain, toUpdateRecord)
			continue
		}
		if toUpdateRecord.ID > 0 {
			// update the record directly. skip the extra search
			logWarningf("%s: Unable to directly update records yet. Record: %+v", domain.Domain, toUpdateRecord)
			continue
		}

		var currentIP string
		if toUpdateRecord.Type == "A" {
			currentIP = ipv4.String()
		} else if ipv6 == nil || ipv6.To4() != nil {
			if ipv6 == nil {
				ipv6 = ipv4
			}

			logWarningf("%s: You are trying to update an IPv6 AAAA record without an IPv6 address: ip: %s config: %+v",
				domain.Domain,
				ipv6,
				toUpdateRecord,
			)
			if config.AllowIPv4InIPv6 {
				currentIP = toIPv6String(ipv6)
				log.Printf("%s: Converting IPv4 `%s` to IPv6 `%s`", domain.Domain, ipv6.String(), currentIP)
			} else {
				continue
			}
		} else {
			currentIP = ipv6.String()
		}

		log.Printf("%s: trying to update `%s` : `%s`", domain.Domain, toUpdateRecord.Type, toUpdateRecord.Name)
		for _, doRecord := range doRecords {
			//log.Printf("%s: checking `%s` : `%s`", domain.Domain, doRecord.Type, doRecord.Name)
			if doRecord.Name == toUpdateRecord.Name && doRecord.Type == toUpdateRecord.Type {
				if doRecord.Data == currentIP && (toUpdateRecord.TTL < 30 || doRecord.TTL == toUpdateRecord.TTL) {
					log.Printf("%s: IP/TTL did not change %+v", domain.Domain, doRecord)
					continue
				}
				log.Printf("%s: updating %+v", domain.Domain, doRecord)
				// set the IP address
				doRecord.Data = currentIP
				if toUpdateRecord.TTL >= 30 && doRecord.TTL != toUpdateRecord.TTL {
					doRecord.TTL = toUpdateRecord.TTL
				}
				update, err := json.Marshal(doRecord)
				checkError(err)
				client := &http.Client{}
				request, err := http.NewRequest("PUT",
					"https://api.digitalocean.com/v2/domains/"+url.PathEscape(domain.Domain)+"/records/"+strconv.FormatInt(int64(doRecord.ID), 10),
					bytes.NewBuffer(update))
				checkError(err)
				request.Header.Set("Content-Type", "application/json")
				request.Header.Add("Authorization", "Bearer "+config.APIKey)
				response, err := client.Do(request)
				checkError(err)
				defer response.Body.Close()
				body, err := ioutil.ReadAll(response.Body)
				checkError(err)
				log.Printf("%s: DO update response for %s: %s", domain.Domain, doRecord.Name, string(body))
				updated++
			}
		}

	}
	log.Printf("%s: %d of %d records updated", domain.Domain, updated, len(domain.Records))
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

// func areZero(bs []byte) bool {
// 	for _, b := range bs {
// 		if b != 0 {
// 			return false
// 		}
// 	}
// 	return true
// }

func main() {
	config = GetConfig()
	currentIPv4, currentIPv6 := CheckLocalIPs()
	if currentIPv4 == nil && currentIPv6 == nil {
		logError("Current IP addresses are not valid, or both are disabled in the config. Check your configuration and internet connection.")
	}

	for _, domain := range config.Domains {
		log.Printf("%s: START", domain.Domain)
		UpdateRecords(domain, currentIPv4, currentIPv6)
		log.Printf("%s: END", domain.Domain)
	}
}
