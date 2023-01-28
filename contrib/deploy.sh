#!/usr/bin/env bash

set -e

DEVENV=${OF_DEV_ENV:-kind}

echo ""
echo "Applying namespaces"
kubectl --context "kind-$DEVENV" apply -f ./namespaces.yml

echo ""
echo "Waiting for helm install to complete."

helm upgrade \
    --kube-context "kind-$DEVENV" \
    --install \
    openfaas \
    ./chart/openfaas \
    --namespace openfaas

kubectl --context "kind-$DEVENV" rollout status deploy/prometheus -n openfaas
kubectl --context "kind-$DEVENV" rollout status deploy/gateway -n openfaas
