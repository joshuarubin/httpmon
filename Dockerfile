FROM golang:alpine as build
RUN apk add --no-cache git gcc musl-dev
COPY . /go/src/jrubin.io/httpmon
WORKDIR /go/src/jrubin.io/httpmon
ENV GO111MODULE on
RUN go build -v

FROM alpine:latest
MAINTAINER Joshua Rubin <joshua@rubixconsulting.com>
ENTRYPOINT ["httpmon"]
COPY --from=build /go/src/jrubin.io/httpmon/httpmon /usr/local/bin/httpmon
