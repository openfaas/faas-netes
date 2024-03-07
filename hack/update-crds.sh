#!/bin/bash

GOPATH=$(go env GOPATH)
controllergen="$GOPATH/bin/controller-gen"
PKG=sigs.k8s.io/controller-tools/cmd/controller-gen@v0.14.0

if [ ! -e "$controllergen" ]; then
  echo "Getting $PKG"
  go install $PKG
fi

"$controllergen" \
  crd \
  schemapatch:manifests=./artifacts/crds \
  paths=./pkg/apis/... \
  output:dir=./artifacts/crds
