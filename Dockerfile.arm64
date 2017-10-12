FROM alexellis2/golang:1.9-arm64 as build
#FROM golang:1.7.5
ENV GOPATH=/go/
RUN mkdir -p /go/src/github.com/openfaas/faas-netes/

WORKDIR /go/src/github.com/openfaas/faas-netes

COPY vendor     vendor
COPY handlers	handlers
COPY types      types
COPY server.go  .

RUN gofmt -l -d $(find . -type f -name '*.go' -not -path "./vendor/*") \
  && CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o faas-netes .

FROM debian:stretch
#FROM alpine:3.5
#RUN apk --no-cache add ca-certificates
RUN apt-get update -qy && apt-get -qy install ca-certificates
WORKDIR /root/

EXPOSE 8080
ENV http_proxy      ""
ENV https_proxy     ""

COPY --from=0 /go/src/github.com/openfaas/faas-netes/faas-netes    .

CMD ["./faas-netes"]
