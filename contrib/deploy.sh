#!/usr/bin/env bash

DEVENV=${OF_DEV_ENV:-kind}

export KUBECONFIG="$(kind get kubeconfig-path --name="$DEVENV")"

echo "Applying namespaces"
kubectl apply -f ./namespaces.yml

sha_cmd="sha256sum"
if [ ! -x "$(command -v $sha_cmd)" ]; then
    sha_cmd="shasum -a 256"
fi

if [ -x "$(command -v $sha_cmd)" ]; then
    sha_cmd="shasum"
fi

PASSWORD=$(head -c 16 /dev/urandom| $sha_cmd | cut -d " " -f 1)
echo -n $PASSWORD > password.txt

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
