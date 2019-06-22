FROM golang:1.10

RUN curl -sLSf https://raw.githubusercontent.com/teamserverless/license-check/master/get.sh | sh
RUN mv ./license-check /usr/bin/license-check && chmod +x /usr/bin/license-check

WORKDIR /go/src/github.com/openfaas/faas-netes
COPY . .

RUN license-check -path /go/src/github.com/openfaas/faas-netes/ --verbose=false "Alex Ellis" "OpenFaaS Author(s)"
RUN gofmt -l -d $(find . -type f -name '*.go' -not -path "./vendor/*") \
    && go test ./test/ \
    && VERSION=$(git describe --all --exact-match `git rev-parse HEAD` | grep tags | sed 's/tags\///') \
    && GIT_COMMIT=$(git rev-list -1 HEAD) \
    && CGO_ENABLED=0 GOOS=linux go build --ldflags "-s -w \
        -X github.com/openfaas/faas-netes/version.GitCommit=${GIT_COMMIT}\
        -X github.com/openfaas/faas-netes/version.Version=${VERSION}" \
        -a -installsuffix cgo -o faas-netes .

FROM alpine:3.9
RUN apk --no-cache add ca-certificates
WORKDIR /root/

EXPOSE 8080
ENV http_proxy      ""
ENV https_proxy     ""

COPY --from=0 /go/src/github.com/openfaas/faas-netes/faas-netes    .

CMD ["./faas-netes"]
