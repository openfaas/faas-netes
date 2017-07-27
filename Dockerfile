FROM golang:1.7.5

RUN mkdir -p /go/src/github.com/alexellis/faas-netes/

WORKDIR /go/src/github.com/alexellis/faas-netes

COPY vendor     vendor
COPY handlers	handlers
COPY types      types
COPY server.go  .

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o faas-netes .

# FROM alpine:latest
# WORKDIR /root/

EXPOSE 8080
ENV http_proxy      ""
ENV https_proxy     ""

# COPY --from=0 /go/src/github.com/alexellis/faas-netes/faas-netes    .

CMD ["./faas-netes"]
