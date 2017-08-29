FROM alpine:3.6

COPY . /go/src/github.com/bobrik/zoidberg

RUN apk --update add go libc-dev ca-certificates && \
    GOPATH=/go go install -v github.com/bobrik/zoidberg/cmd/... && \
    apk del go libc-dev

ENTRYPOINT ["/go/bin/zoidberg"]
