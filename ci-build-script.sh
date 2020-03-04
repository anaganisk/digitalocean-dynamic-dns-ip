#!/bin/bash
package=$1
if [[ -z "$package" ]]; then
  echo "usage: $0 <package-name>"
  exit 1
fi
package_split=(${package//\// })
package_name=${package_split[-1]}
#see https://en.wikipedia.org/wiki/Raspberry_Pi#Specifications for RPi arm versions
#see https://github.com/golang/go/wiki/GoArm
platforms=("windows/amd64" "windows/386" "darwin/amd64" "linux/386" "linux/amd64" "linux/arm/6" "linux/arm/7" "linux/arm64/8" "freebsd/386" "freebsd/amd64")
mkdir -p releases
for platform in "${platforms[@]}"
do
    platform_split=(${platform//\// })
    GOOS=${platform_split[0]}
    GOARCH=${platform_split[1]}
    GOARM=${platform_split[2]}
    output_name='digitalocean-dynamic-dns-ip-'$GOOS'-'$GOARCH
    if [[ ! -z "$GOARM" ]]; then
        output_name+="v${GOARM}"
        # arm v8 is only supported for ARCH=arm64 and requires an empty GOARM version
        if [[ "8" -eq "$GOARM" ]]; then
            GOARM=""
        fi
    fi
    if [ $GOOS = "windows" ]; then
        output_name+='.exe'
    fi
    env GOOS="$GOOS" GOARCH="$GOARCH" GOARM="$GOARM" go build -o "releases/$output_name" "$package_name"
    if [ $? -ne 0 ]; then
        echo 'An error has occurred! Aborting the script execution...'
        exit 1
    fi
done