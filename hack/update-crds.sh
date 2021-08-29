#!/bin/bash

export controllergen="$GOPATH/bin/controller-gen"
export PKG=sigs.k8s.io/controller-tools/cmd/controller-gen@v0.6.2

if [ ! -e "$gen" ]
then
echo "Getting $PKG"
    go get $PKG
fi

"$controllergen" \
  crd \
  schemapatch:manifests=./artifacts/crds \
  paths=./pkg/apis/... \
  output:dir=./artifacts/crds
