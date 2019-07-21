FROM golang:1.12 as builder

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

COPY --from=builder /usr/bin/node-labels-to-files /usr/bin/
