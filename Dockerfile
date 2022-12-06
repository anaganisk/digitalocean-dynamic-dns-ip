FROM docker.io/library/golang:1.19-bullseye as dev

FROM dev as intermediate

COPY . /build
WORKDIR /build
RUN go build -o doddns && \
    ldd doddns | tr -s '[:blank:]' '\n' | grep '^/' | \
    awk '{printf("%s\n", $1); system("readlink -f " $1)}' | \
    xargs -I % sh -c 'mkdir -p $(dirname deps%); cp % deps%;'

FROM scratch as prod

COPY --from=intermediate /build/deps /
COPY --from=intermediate /build/doddns /doddns
COPY --from=intermediate /etc/ssl/certs/* /etc/ssl/certs/
COPY --from=intermediate /usr/share/ca-certificates /usr/share/ca-certificates

WORKDIR /
ENTRYPOINT ["/doddns"]
