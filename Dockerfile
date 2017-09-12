FROM golang:1.7.5

RUN mkdir -p /go/src/github.com/alexellis/faas-netes/

WORKDIR /go/src/github.com/alexellis/faas-netes

COPY vendor     vendor
COPY handlers	handlers
COPY types      types
COPY server.go  .

RUN gofmt -l -d $(find . -type f -name '*.go' -not -path "./vendor/*") \
  && CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o faas-netes .

FROM alpine:3.5
RUN apk --no-cache add ca-certificates
WORKDIR /root/

EXPOSE 8080
ENV http_proxy      ""
ENV https_proxy     ""

COPY --from=0 /go/src/github.com/alexellis/faas-netes/faas-netes    .

CMD ["./faas-netes"]
