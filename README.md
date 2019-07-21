# DIGITAL OCEAN DYNAMIC IP API CLIENT
A simple script in Go language to automatically update Digital ocean DNS records if you have a dynamic IP. Since it can be compiled on any platform, you can use it along with raspberrypi etc.

To find your Dynamic IP, this program will call out to https://diagnostic.opendns.com/myip.

## requirements
Requires Git and Go for building.

Requires that the record already exists in DigitalOcean's DNS so that it can be updated.
(manually find your IP and add it to DO's DNS it will later be updated)

Requires a Digital Ocean API key that can be created at https://cloud.digitalocean.com/account/api/tokens.

## Usage
```bash
git clone https://github.com/anaganisk/digitalocean-dynamic-dns-ip.git
```
create a file ".digitalocean-dynamic-ip.json"(dot prefix to hide the file) and place it user home directory and add the following json

```json
{
  "apikey": "samplekeydasjkdhaskjdhrwofihsamplekey",
  "doPageSize" : 20,
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
The TTL can optionally be updated if passed in the configuration. Digital Ocean has a minimum TTL of 30 seconds. The `type` and the `name` must match existing records in the Digital Ocean DNS configuration. Only `types` of `A` and `AAAA` allowed at the moment.

If you want to reduce the number of calls made to the digital ocean API and have more than 20 DNS records in your domain, you can adjust the `doPageSize` parameter. By default, Digital Ocean returns 20 records per page.

```bash
#run
go build digitalocean-dynamic-ip.go
./digitalocean-dynamic-ip
```

Optionally, you can create the configuration file with any name wherever you want, and pass it as a command line argument:
````bash
#run
./digitalocean-dynamic-ip /path/tp/my/config.json
```

You can either set this to run periodically with a cronjob or use your own method.
```bash
# run crontab -e
# sample cron job task 

# m h  dom mon dow   command
*/5 * * * * /home/user/digitalocean-dynamic-dns-ip/digitalocean-dynamic-ip
```
