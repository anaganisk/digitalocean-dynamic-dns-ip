# DIGITAL OCEAN DYNAMIC IP API CLIENT
A simple script in Go language to automatically update Digital ocean DNS records if you have a dynamic IP. Since it can be compiled on any platform, you can use it along with raspberrypi etc.

## requirements
Requires Git and Go for building.

Requires that the record already exists in DigitalOcean's DNS.

## Usage
```bash
git clone https://github.com/anaganisk/digitalocean-dynamic-dns-ip.git
```
create a file ".digitalocean-dynamic-ip.json"(dot prefix to hide the file) and place it user home directory and add the following json

```json
{
  "apikey": "samplekeydasjkdhaskjdhrwofihsamplekey",
  "domain": "example.com",
  "records": [
    {
      "name": "subdomain",
      "type": "A"
    }
  ]
}
```
```bash
#run
go build digitalocean-dynamic-ip.go
./digitalocean-dynamic-ip
```
You can either set this to run periodically with a cronjob or use your own method.
```bash
# run crontab -e
# sample cron job task 

# m h  dom mon dow   command
*/5 * * * * /home/user/digitalocean-dynamic-dns-ip/digitalocean-dynamic-ip
```
