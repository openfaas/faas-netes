#!/bin/bash

export KUBECONFIG="$(kind get kubeconfig-path)"

kubectl -n kube-system create sa tiller \
 && kubectl create clusterrolebinding tiller \
      --clusterrole cluster-admin \
      --serviceaccount=kube-system:tiller

helm init --skip-refresh --upgrade --service-account tiller --wait
