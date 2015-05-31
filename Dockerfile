FROM alpine:3.2

COPY . /go/src/github.com/bobrik/zoidberg

RUN apk --update add go && \
    GOPATH=/go:/go/src/github.com/bobrik/zoidberg/cmd/marathon-explorer/Godeps/_workspace go get github.com/bobrik/zoidberg/cmd/marathon-explorer && \
    GOPATH=/go:/go/src/github.com/bobrik/zoidberg/cmd/marathon-explorer/Godeps/_workspace go get github.com/bobrik/zoidberg/cmd/mesos-explorer && \
    apk del go
