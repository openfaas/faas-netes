#!/bin/bash

export KUBECONFIG="$(kind get kubeconfig-path)"

kubectl -n kube-system create sa tiller \
 && kubectl create clusterrolebinding tiller \
      --clusterrole cluster-admin \
      --serviceaccount=kube-system:tiller

helm init --skip-refresh --upgrade --service-account tiller

for i in {0...120};
do
    sleep 1
    echo "Checking if the tiller deployment is ready"
    ready=$(kubectl get deploy/tiller-deploy -n kube-system --output="jsonpath={.status.availableReplicas}")
    if [ "$ready" == "1" ]; then
        echo "Tiller ready"
        break
    fi
done
