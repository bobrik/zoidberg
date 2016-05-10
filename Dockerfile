FROM alpine:3.3

COPY . /go/src/github.com/bobrik/zoidberg

RUN apk --update add go ca-certificates && \
    export GOPATH=/go GO15VENDOREXPERIMENT=1 && \
    go get github.com/bobrik/zoidberg/... && \
    apk del go

ENTRYPOINT ["/go/bin/zoidberg"]
