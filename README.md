# DIGITAL OCEAN DYNAMIC IP API CLIENT 

A simple script in Go language to automatically update Digital ocean DNS records if you have a dynamic IP. Since it can be compiled on any platform, you can use it along with raspberrypi etc.

To find your Dynamic IP, this program will call out to https://api.ipify.org/ for ipv4 addresses and https://api64.ipify.org/ for ipv6 addresses. This is to support dual-stack environments. (These URLs can be customized; see Usage, below.)

## Requirements

- The record must already exist in DigitalOcean's DNS so that it can be updated.
  (manually find your IP and add it to DO's DNS it will later be updated)
- A Digital Ocean API key that can be created at https://cloud.digitalocean.com/account/api/tokens.

## Installation

Download the prebuilt binaries from [releases](https://github.com/anaganisk/digitalocean-dynamic-dns-ip/releases),

or Build from source

```bash
# Requires Git, Go 1.8+(GO 1.11 if you want to use GO111MODULE=on).
# clone the repo in ~/go/src/github.com/anaganisk:
git clone https://github.com/anaganisk/digitalocean-dynamic-dns-ip.git
# Skip to next step, if you have GO111MODULE=on in your environment it is fetched automatically
go get github.com/mitchellh/go-homedir
# build the project
go build
```

## Usage

Create a file `.digitalocean-dynamic-ip.json` (dot prefix to hide the file) and place it in your home directory. Add the following JSON (or refer to the sample configuration file, `digitalocean-dynamic-ip.sample.json`):

```json
{
  "apikey": "samplekeydasjkdhaskjdhrwofihsamplekey",
  "doPageSize": 20,
  "useIPv4": true,
  "useIPv6": false,
  "allowIPv4InIPv6": false,
  "ipv4CheckUrl": "https://api.ipify.org/?format=text",
  "domains": [
    {
      "domain": "example.com",
      "records": [
        {
          "name": "subdomainOrRecord",
          "type": "A"
        }
      ]
    },
    {
      "domain": "example2.com",
      "records": [
        {
          "name": "subdomainOrRecord2",
          "type": "A",
          "TTL": 30
        }
      ]
    }
  ]
}
```

The TTL can optionally be updated if passed in the configuration. Digital Ocean has a minimum TTL of 30 seconds. The `type` and the `name` must match existing records in the Digital Ocean DNS configuration. Only `types` of `A` and `AAAA` allowed at the moment. Use the `name` of `"@"` to update a the top level domain record itself instead of a sub domain.

If you want to reduce the number of calls made to the digital ocean API and have more than 20 DNS records in your domain, you can adjust the `doPageSize` parameter. By default, Digital Ocean returns 20 records per page. Digital Ocean has a max page size of 200 items.

By default, the configuration checks both IPv4 and IPv6 addresses assuming your provider set up your connection as dual stack. If you know you only have ipv4 or ipv6 you can disable using one or the other in the config. To disable one or the other, set the `useIPv4` or `useIPv6` settings to `false`. If the options aren't present, or are set to `null`, then the configuration assumes a value of `true`.

The `allowIPv4InIPv6` configuration option will allow adding an IPv4 address to be used in a AAAA record for IPv6 lookups.

The `ipv4CheckUrl` and `ipv6CheckUrl` configuration settings are optional. If set, they must be URLs which respond to a GET request, with a plaintext response containing only your IP address. If unset, they default to `https://api.ipify.org/?format=text` and `https://api64.ipify.org/?format=text` respectively.

```bash
# after running `go build digitalocean-dynamic-ip.go`, run:
./digitalocean-dynamic-ip
```

Optionally, you can create the configuration file with any name wherever you want, and pass it as a command line argument:

```bash
#run:
./digitalocean-dynamic-ip /path/to/my/config.json
```

You can either set this to run periodically with a cronjob or use your own method.

```bash
# run `crontab -e` to edit your crontab
# sample cron job task

# m h  dom mon dow   command
*/5 * * * * /home/user/digitalocean-dynamic-dns-ip/digitalocean-dynamic-ip
```
