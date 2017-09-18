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
		panic(err)
	}
}

// ClientConfig : configuration json
type ClientConfig struct {
	APIKey string      `json:"apiKey"`
	Domain string      `json:"domain"`
	Record []DNSRecord `json:"records"`
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

func main() {
	homeDirectory, err := homedir.Dir()
	checkError(err)
	getfile, err := ioutil.ReadFile(homeDirectory + "/.digitalocean-dynamic-ip.json")
	checkError(err)
	var config ClientConfig
	json.Unmarshal(getfile, &config)
	checkError(err)

	// check current local ip

	currentIPRequest, err := http.Get("https://diagnostic.opendns.com/myip")
	checkError(err)
	defer currentIPRequest.Body.Close()
	currentIPRequestParse, err := ioutil.ReadAll(currentIPRequest.Body)
	checkError(err)
	currentIP := string(currentIPRequestParse)

	// get current dns record ip

	client := &http.Client{}
	request, err := http.NewRequest("GET",
		"https://api.digitalocean.com/v2/domains/"+string(config.Domain)+"/records",
		nil)
	checkError(err)
	request.Header.Add("Content-type", "Application/json")
	request.Header.Add("Authorization", "Bearer "+string(config.APIKey))
	response, err := client.Do(request)
	checkError(err)
	defer response.Body.Close()
	body, err := ioutil.ReadAll(response.Body)
	var jsonResponse DOResponse
	e := json.Unmarshal(body, &jsonResponse)
	checkError(e)

	// update ip by matching dns records and config

	for _, record := range jsonResponse.DomainRecords {
		for _, configRecord := range config.Record {
			if configRecord.Name == record.Name && configRecord.Type == record.Type && currentIP != record.Data {
				update := []byte(`{"type":"` + configRecord.Type + `","data":"` + currentIP + `"}`)
				client := &http.Client{}
				request, err := http.NewRequest("PUT",
					"https://api.digitalocean.com/v2/domains/"+string(config.Domain)+"/records/"+strconv.FormatInt(int64(record.ID), 10),
					bytes.NewBuffer(update))
				checkError(err)
				request.Header.Set("Content-Type", "application/json")
				request.Header.Add("Authorization", "Bearer "+string(config.APIKey))
				response, err := client.Do(request)
				checkError(err)
				defer response.Body.Close()
				body, err := ioutil.ReadAll(response.Body)
				log.Printf("DO update response for %s: %s", record.Name, string(body))
			}
		}
	}
}
