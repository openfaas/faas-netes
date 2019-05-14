#!/bin/bash

echo "go get -u sigs.k8s.io/kind"

export GO111MODULE="on"
go get -u sigs.k8s.io/kind@0a6ceae
