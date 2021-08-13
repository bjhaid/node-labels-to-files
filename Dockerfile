FROM golang:1.14 as builder

ENV GOPATH=/go
ENV GO111MODULE=on

RUN go get honnef.co/go/tools/cmd/staticcheck

ENV GOFLAGS=-mod=vendor
COPY . /go/src/github.com/bjhaid/node-labels-to-files
WORKDIR /go/src/github.com/bjhaid/node-labels-to-files

RUN go test \
  && staticcheck . \
  && go build -o /usr/bin/node-labels-to-files

FROM debian:buster-slim

RUN apt-get update -y && \
    apt-get dist-upgrade -u -y && \
    rm -rf /var/lib/apt/lists/*

COPY --from=builder /usr/bin/node-labels-to-files /usr/bin/
