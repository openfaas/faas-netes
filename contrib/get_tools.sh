#!/bin/bash

set -e

KIND_VERSION="v0.9.0"
# Causes a validation failure when linting due to CRDs moving to v1
# HELM_VERSION="v3.4.0"
HELM_VERSION="v3.0.3"
KUBE_VERSION="v1.18.8"
ARKADE_VERSION="0.6.21"

echo "Downloading arkade"

curl -SLs https://github.com/alexellis/arkade/releases/download/$ARKADE_VERSION/arkade > arkade
chmod +x ./arkade


if [[ "$1" ]]; then
  KUBE_VERSION=$1
fi

./arkade get kind --version $KIND_VERSION
./arkade get helm --version $HELM_VERSION
./arkade get kubectl --version $KUBE_VERSION
./arkade get faas-cli

sudo mv $HOME/.arkade/bin/* /usr/local/bin/
