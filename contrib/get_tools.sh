#!/bin/bash

set -e

KIND_VERSION="v0.22.0"
# Causes a validation failure when linting due to CRDs moving to v1
# HELM_VERSION="v3.4.0"
HELM_VERSION="v3.14.4"
KUBE_VERSION="v1.29.2"
ARKADE_VERSION="0.11.9"

echo "Downloading arkade"

curl -SLs https://github.com/alexellis/arkade/releases/download/$ARKADE_VERSION/arkade > arkade
chmod +x ./arkade


if [[ "$1" ]]; then
  KUBE_VERSION=$1
fi

./arkade get kind@$KIND_VERSION \
  helm@$HELM_VERSION \
  kubectl@$KUBE_VERSION \
  faas-cli

sudo mv $HOME/.arkade/bin/* /usr/local/bin/
