#!/usr/bin/env bash

DEVENV=${OF_DEV_ENV:-kind}

export KUBECONFIG="$(kind get kubeconfig-path --name="$DEVENV")"

kubectl -n kube-system create sa tiller \
 && kubectl create clusterrolebinding tiller \
      --clusterrole cluster-admin \
      --serviceaccount=kube-system:tiller

helm init --skip-refresh --upgrade --service-account tiller --wait

# wait, helm/tiller to finish
kubectl rollout status deploy tiller-deploy --watch -n kube-system
