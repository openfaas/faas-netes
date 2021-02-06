#!/bin/bash

if [[ ! -v BUILDPLATFORM ]]; then
    CGO_ENABLED=0 go test -v ./...
elif [[ "$BUILDPLATFORM" == "$TARGETPLATFORM" ]]; then
    echo "testing in docker"
    CGO_ENABLED=${CGO_ENABLED} GOOS=${TARGETOS} GOARCH=${TARGETARCH} go test -v ./...
else
    echo "skipping tests during cross builds"
fi