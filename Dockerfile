FROM ghcr.io/openfaas/license-check:0.4.2 as license-check

FROM --platform=${BUILDPLATFORM:-linux/amd64} golang:1.19 as build

ARG TARGETPLATFORM
ARG BUILDPLATFORM
ARG TARGETOS
ARG TARGETARCH

ARG VERSION
ARG GIT_COMMIT

ENV CGO_ENABLED=0
ENV GO111MODULE=on
ENV GOFLAGS=-mod=vendor

COPY --from=license-check /license-check /usr/bin/

WORKDIR /go/src/github.com/openfaas/faas-netes
COPY . .

RUN license-check -path /go/src/github.com/openfaas/faas-netes/ --verbose=false "Alex Ellis" "OpenFaaS Author(s)"
RUN gofmt -l -d $(find . -type f -name '*.go' -not -path "./vendor/*")
RUN CGO_ENABLED=${CGO_ENABLED} GOOS=${TARGETOS} GOARCH=${TARGETARCH} go test -v ./...

RUN GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build \
        --ldflags "-s -w \
        -X github.com/openfaas/faas-netes/version.GitCommit=${GIT_COMMIT}\
        -X github.com/openfaas/faas-netes/version.Version=${VERSION}" \
        -o faas-netes .

FROM --platform=${TARGETPLATFORM:-linux/amd64} alpine:3.17 as ship
LABEL org.label-schema.license="MIT" \
      org.label-schema.vcs-url="https://github.com/openfaas/faas-netes" \
      org.label-schema.vcs-type="Git" \
      org.label-schema.name="openfaas/faas-netes" \
      org.label-schema.vendor="openfaas" \
      org.label-schema.docker.schema-version="1.0"

RUN apk --no-cache add \
    ca-certificates

RUN addgroup -S app \
    && adduser -S -g app app

WORKDIR /home/app

EXPOSE 8080

ENV http_proxy      ""
ENV https_proxy     ""

COPY --from=build /go/src/github.com/openfaas/faas-netes/faas-netes    .
RUN chown -R app:app ./

USER app

CMD ["./faas-netes"]
