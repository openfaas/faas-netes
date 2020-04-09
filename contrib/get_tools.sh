#!/bin/bash

set -e

KIND_VERSION="v0.7.0"
HELM_VERSION=v3.0.3
KUBE_VERSION=v1.17.0

if [[ "$1" ]]; then
  KUBE_VERSION=$1
fi

echo ">>> Installing kind $KIND_VERSION"
curl -sSLfo kind "https://github.com/kubernetes-sigs/kind/releases/download/$KIND_VERSION/kind-linux-amd64" && \
chmod +x kind && \
sudo mv kind /usr/local/bin/kind

echo ">>> Installing kubectl"
OS=linux
curl -sSLfo kubectl https://storage.googleapis.com/kubernetes-release/release/$KUBE_VERSION/bin/$OS/amd64/kubectl && \
chmod +x kubectl && \
sudo mv kubectl /usr/local/bin/

echo ">>> Installing Helm v3"
curl -sSLf https://get.helm.sh/helm-${HELM_VERSION}-linux-amd64.tar.gz | tar xz && \
sudo mv linux-amd64/helm /usr/local/bin/ && \
rm -rf linux-amd64

echo ">>> Installing faas-cli"
curl -sSLf https://cli.openfaas.com | sudo sh
faas-cli version
