FROM golang:1.9.2

RUN mkdir -p /go/src/github.com/openfaas/faas-netes/

WORKDIR /go/src/github.com/openfaas/faas-netes

COPY vendor     vendor
COPY handlers	handlers
COPY types      types
COPY server.go  .

RUN curl -sL https://github.com/alexellis/license-check/releases/download/0.2.2/license-check > /usr/bin/license-check \
    && chmod +x /usr/bin/license-check
RUN license-check -path ./ --verbose=false "Alex Ellis" "OpenFaaS Project"
RUN gofmt -l -d $(find . -type f -name '*.go' -not -path "./vendor/*") \
  && CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o faas-netes .

FROM alpine:3.6
RUN apk --no-cache add ca-certificates
WORKDIR /root/

EXPOSE 8080

ENV http_proxy      ""
ENV https_proxy     ""

COPY --from=0 /go/src/github.com/openfaas/faas-netes/faas-netes    .

CMD ["./faas-netes"]
