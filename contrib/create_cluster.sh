#!/usr/bin/env bash

DEVENV=${OF_DEV_ENV:-kind-kind}
KUBE_VERSION=v1.17.0

echo ">>> Creating Kubernetes ${KUBE_VERSION} cluster ${DEVENV}"

kind create cluster --wait 5m --image kindest/node:${KUBE_VERSION} --name "$DEVENV"

echo ">>> Waiting for CoreDNS"
kubectl --context "$DEVENV" -n kube-system rollout status deployment/coredns
