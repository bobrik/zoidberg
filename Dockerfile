FROM alpine:3.1

RUN apk --update add go git

COPY . /go/src/github.com/bobrik/zoidberg

RUN GOPATH=/go go get github.com/bobrik/zoidberg/...
