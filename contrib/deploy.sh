#!/bin/bash

export KUBECONFIG="$(kind get kubeconfig-path)"

echo "Applying namespaces"
kubectl apply -f ./namespaces.yml

echo $(head -c 16 /dev/urandom| sha256sum | cut -d " " -f 1) > ./password.txt # Store password in password.txt file
PASSWORD=$(cat ./password.txt)

kubectl -n openfaas create secret generic basic-auth \
--from-literal=basic-auth-user=admin \
--from-literal=basic-auth-password="$PASSWORD"

echo "Waiting for helm install to complete."

helm install ./chart/openfaas \
    --name openfaas \
    --namespace openfaas  \
    --set basic_auth=true \
    --set functionNamespace=openfaas-fn \
    --wait

