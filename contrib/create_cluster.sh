#!/usr/bin/env bash

DEVENV=${OF_DEV_ENV:-kind}
KUBE_VERSION=v1.17.0

echo ">>> Creating Kubernetes ${KUBE_VERSION} cluster"
kind create cluster --wait 5m --image kindest/node:${KUBE_VERSION} --name "$DEVENV"

export KUBECONFIG="$(kind get kubeconfig-path --name="$DEVENV")"

echo ">>> Waiting for CoreDNS"
kubectl -n kube-system rollout status deployment/coredns
