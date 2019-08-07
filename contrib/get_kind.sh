#!/bin/bash

echo "go get -u sigs.k8s.io/kind@v0.4.0"

export GO111MODULE="on"
go get -u sigs.k8s.io/kind@v0.4.0
