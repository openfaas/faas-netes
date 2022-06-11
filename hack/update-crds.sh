#!/bin/bash

export controllergen="$GOBIN/controller-gen"
export PKG=sigs.k8s.io/controller-tools/cmd/controller-gen@v0.6.2

if [ ! -e "$controllergen" ]; then
  echo "Getting $PKG"
  go install $PKG
fi

"$controllergen" \
  crd \
  schemapatch:manifests=./artifacts/crds \
  paths=./pkg/apis/... \
  output:dir=./artifacts/crds
