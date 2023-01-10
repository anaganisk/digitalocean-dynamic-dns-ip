# This version does not run as a "FROM scratch" container.
# Useful on systems where the other version does not work for unknown reasons.

FROM docker.io/library/golang:1.19-bullseye as dev

FROM dev as intermediate

COPY . /build
WORKDIR /build
RUN go build -o /doddns 

WORKDIR /
ENTRYPOINT ["/doddns"]
