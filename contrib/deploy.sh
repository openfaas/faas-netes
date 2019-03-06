#!/bin/bash

export KUBECONFIG="$(kind get kubeconfig-path)"

echo "Applying namespaces"
kubectl apply -f ./namespaces.yml

PASSWORD="something_random"
kubectl -n openfaas create secret generic basic-auth \
--from-literal=basic-auth-user=admin \
--from-literal=basic-auth-password="$PASSWORD"

helm template ./chart/openfaas \
    --name openfaas \
    --namespace openfaas  \
    --set basic_auth=true \
    --set functionNamespace=openfaas-fn > ./openfaas.yaml

kubectl apply -f ./openfaas.yaml