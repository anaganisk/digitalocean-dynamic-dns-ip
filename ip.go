package main

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"sync"

	"github.com/anaganisk/digitalocean-dynamic-dns-ip/config"
	"github.com/anaganisk/digitalocean-dynamic-dns-ip/constants"
	"github.com/anaganisk/digitalocean-dynamic-dns-ip/logger"
)

// CheckPublicIPs : get current IP of server. checks both IPv4 and Ipv6 to support dual stack environments
func CheckPublicIPs() (ipv4, ipv6 net.IP) {
	ipv4CheckURL := constants.DefaultIPv4CheckURL
	ipv6CheckURL := constants.DefaultIPv6CheckURL
	conf := config.Get()
	if len(conf.IPv4CheckURL) > 0 {
		ipv4CheckURL = conf.IPv4CheckURL
	}
	if len(conf.IPv6CheckURL) > 0 {
		ipv6CheckURL = conf.IPv6CheckURL
	}

	var wg sync.WaitGroup

	if conf.UseIPv4 == nil || *(conf.UseIPv4) {
		wg.Add(1)
		go func(url string) {
			defer wg.Done()
			ip, err := getIp("IPv4", url)
			if err != nil {
				logger.Warning(err.Error())
				return
			}
			ipv4 = ip
		}(ipv4CheckURL)
	}

	if conf.UseIPv6 == nil || *(conf.UseIPv6) {
		wg.Add(1)
		go func(url string) {
			defer wg.Done()
			ip, err := getIp("IPv6", url)
			if err != nil {
				logger.Warning(err.Error())
				return
			}
			ipv6 = ip
		}(ipv6CheckURL)
	}
	wg.Wait()

	return ipv4, ipv6
}

func getIp(ipType, url string) (net.IP, error) {
	logger.Debug("Checking %s with URL: %s", ipType, url)

	ipString, err := getURLBody(url)
	if err != nil {
		return nil, err
	}

	if ipString != "" {
		ip := net.ParseIP(ipString)
		if ip != nil {
			if strings.ToLower(ipType) == "ipv4" {
				ip = ip.To4() // ensure IPv4 when expected
			}
			logger.Debug("Discovered %s address `%s`", ipType, ip.String())
		}
		if ip == nil {
			return nil, fmt.Errorf("unable to parse %q as a %s address", ipString, ipType)
		}
		return ip, nil
	}

	return nil, fmt.Errorf(
		"No %s address found. Consider disabling %s checks in the config `\"use%s\": false`",
		ipType, ipType, ipType,
	)
}

// getURLBody fetches the body of the given URL as a string
func getURLBody(url string) (string, error) {
	client := &http.Client{Timeout: config.Get().GetHTTPTimeout()}
	request, err := client.Get(url)
	logger.CheckError(err)
	defer request.Body.Close()
	body, err := io.ReadAll(request.Body)
	logger.CheckError(err)
	return string(body), nil
}
