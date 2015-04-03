FROM alpine:3.1

RUN apk --update add go git

RUN mkdir /go && \
    GOPATH=/go go get github.com/gambol99/go-marathon && \
    GOPATH=/go go get github.com/samuel/go-zookeeper/zk

COPY . /go/src/github.com/bobrik/zoidberg

RUN GOPATH=/go go get github.com/bobrik/zoidberg/...

