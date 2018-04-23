FROM golang:1.9.4

RUN mkdir -p /go/src/github.com/openfaas/faas-netes/

WORKDIR /go/src/github.com/openfaas/faas-netes

COPY . .

RUN curl -sL https://github.com/alexellis/license-check/releases/download/0.2.2/license-check > /usr/bin/license-check \
    && chmod +x /usr/bin/license-check
RUN license-check -path ./ --verbose=false "Alex Ellis" "OpenFaaS Project"
RUN gofmt -l -d $(find . -type f -name '*.go' -not -path "./vendor/*") \
    && VERSION=$(git describe --all --exact-match `git rev-parse HEAD` | grep tags | sed 's/tags\///') \
    && GIT_COMMIT=$(git rev-list -1 HEAD) \
    && CGO_ENABLED=0 GOOS=linux go build --ldflags "-s -w \
        -X github.com/openfaas/faas-netes/version.GitCommit=${GIT_COMMIT}\
        -X github.com/openfaas/faas-netes/version.Version=${VERSION}" \
        -a -installsuffix cgo -o faas-netes .

FROM alpine:3.7

RUN addgroup -S app \
    && adduser -S -g app app \
    && apk --no-cache add \
    ca-certificates
WORKDIR /home/app

EXPOSE 8080

ENV http_proxy      ""
ENV https_proxy     ""

COPY --from=0 /go/src/github.com/openfaas/faas-netes/faas-netes    .
RUN chown -R app:app ./

USER app

CMD ["./faas-netes"]
