#!/usr/bin/env bash

set -o errexit

HELM_VERSION=v3.0.3
KIND_VERSION=v0.7.0
KUBE_VERSION=v1.17.0

if [[ "$1" ]]; then
  KUBE_VERSION=$1
fi

echo ">>> Installing kubectl"
curl -LO https://storage.googleapis.com/kubernetes-release/release/$(curl -s https://storage.googleapis.com/kubernetes-release/release/stable.txt)/bin/linux/amd64/kubectl && \
chmod +x kubectl && \
sudo mv kubectl /usr/local/bin/

echo ">>> Installing kind $VER"

curl -sSLo kind "https://github.com/kubernetes-sigs/kind/releases/download/$KIND_VERSION/kind-linux-amd64"
chmod +x kind
sudo mv kind /usr/local/bin/kind

echo ">>> Creating Kubernetes ${KUBE_VERSION} cluster"
kind create cluster --wait 5m --image kindest/node:${KUBE_VERSION}

echo ">>> Waiting for CoreDNS"
kubectl -n kube-system rollout status deployment/coredns

echo ">>> Installing Helm v3"
curl -sSL https://get.helm.sh/helm-${HELM_VERSION}-linux-amd64.tar.gz | tar xz && \
sudo mv linux-amd64/helm /usr/local/bin/ && \
rm -rf linux-amd64

echo ">>> Add OpenFaaS Helm repo"
helm repo add openfaas https://openfaas.github.io/faas-netes/
