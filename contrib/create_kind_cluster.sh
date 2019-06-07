#!/usr/bin/env bash

DEVENV=${OF_DEV_ENV:-kind}

kind create cluster --name "$DEVENV"

export KUBECONFIG="$(kind get kubeconfig-path --name="$DEVENV")"

# wait, roughly for the cluster to finish starting
kubectl rollout status deploy coredns --watch -n kube-system
