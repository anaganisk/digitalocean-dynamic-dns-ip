package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"

	homedir "github.com/mitchellh/go-homedir"
)

func checkError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

// ClientConfig : configuration json
type ClientConfig struct {
	APIKey  string   `json:"apiKey"`
	Domains []Domain `json:"domains"`
}

// Domain : domains to be changed
type Domain struct {
	Domain  string      `json:"domain"`
	Records []DNSRecord `json:"records"`
}

// DNSRecord : Modifyiable DNS record
type DNSRecord struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
	Data string `json:"data"`
	Type string `json:"type"`
}

// DOResponse : DigitalOcean DNS Records response.
type DOResponse struct {
	DomainRecords []DNSRecord `json:"domain_records"`
}

//GetConfig : get configuration file ~/.digitalocean-dynamic-ip.json
func GetConfig() ClientConfig {
	homeDirectory, err := homedir.Dir()
	checkError(err)
	getfile, err := ioutil.ReadFile(homeDirectory + "/.digitalocean-dynamic-ip.json")
	checkError(err)
	var config ClientConfig
	json.Unmarshal(getfile, &config)
	checkError(err)
	return config
}

//CheckLocalIP : get current IP of server.
func CheckLocalIP() string {
	currentIPRequest, err := http.Get("https://diagnostic.opendns.com/myip")
	checkError(err)
	defer currentIPRequest.Body.Close()
	currentIPRequestParse, err := ioutil.ReadAll(currentIPRequest.Body)
	checkError(err)
	return string(currentIPRequestParse)
}

//GetDomainRecords : Get DNS records of current domain.
func GetDomainRecords(apiKey string, domain string) DOResponse {
	client := &http.Client{}
	request, err := http.NewRequest("GET",
		"https://api.digitalocean.com/v2/domains/"+domain+"/records",
		nil)
	checkError(err)
	request.Header.Add("Content-type", "Application/json")
	request.Header.Add("Authorization", "Bearer "+apiKey)
	response, err := client.Do(request)
	checkError(err)
	defer response.Body.Close()
	body, err := ioutil.ReadAll(response.Body)
	var jsonDOResponse DOResponse
	e := json.Unmarshal(body, &jsonDOResponse)
	checkError(e)
	return jsonDOResponse
}

// UpdateRecords : Update DNS records of domain
func UpdateRecords(apiKey string, domain string, currentIP string, currentRecords DOResponse, toUpdateRecords []DNSRecord) {
	for _, currentRecord := range currentRecords.DomainRecords {
		for _, toUpdateRecord := range toUpdateRecords {
			if toUpdateRecord.Name == currentRecord.Name && toUpdateRecord.Type == currentRecord.Type && currentIP != currentRecord.Data {
				update := []byte(`{"type":"` + toUpdateRecord.Type + `","data":"` + currentIP + `"}`)
				client := &http.Client{}
				request, err := http.NewRequest("PUT",
					"https://api.digitalocean.com/v2/domains/"+domain+"/records/"+strconv.FormatInt(int64(currentRecord.ID), 10),
					bytes.NewBuffer(update))
				checkError(err)
				request.Header.Set("Content-Type", "application/json")
				request.Header.Add("Authorization", "Bearer "+apiKey)
				response, err := client.Do(request)
				checkError(err)
				defer response.Body.Close()
				body, err := ioutil.ReadAll(response.Body)
				log.Printf("DO update response for %s: %s\n", currentRecord.Name, string(body))
			}
		}
	}
}

func main() {
	config := GetConfig()
	currentIP := CheckLocalIP()

	for _, domains := range config.Domains {
		domainName := domains.Domain
		apiKey := config.APIKey
		currentDomainRecords := GetDomainRecords(apiKey, domainName)
		log.Println(domainName)
		UpdateRecords(apiKey, domainName, currentIP, currentDomainRecords, domains.Records)
	}
}
