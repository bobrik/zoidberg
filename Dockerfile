FROM alpine:3.1

RUN apk --update add go

COPY . /go/src/github.com/bobrik/zoidberg

RUN GOPATH=/go:/go/src/github.com/bobrik/zoidberg/cmd/marathon-explorer/Godeps/_workspace go get github.com/bobrik/zoidberg/cmd/marathon-explorer && \
    GOPATH=/go:/go/src/github.com/bobrik/zoidberg/cmd/marathon-explorer/Godeps/_workspace go get github.com/bobrik/zoidberg/cmd/mesos-explorer
