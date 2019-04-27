#!/bin/bash

export KUBECONFIG="$(kind get kubeconfig-path)"

# install again since certifier doesn't work with basic_auth
helm del --purge openfaas
helm install ./chart/openfaas \
    --name openfaas \
    --namespace openfaas  \
    --set basic_auth=false \
    --set functionNamespace=openfaas-fn \
    --wait

kubectl port-forward deploy/gateway -n openfaas 8080:8080 &

export OPENFAAS_URL=http://127.0.0.1:8080/

mkdir -p ./tmp/
cd tmp

git clone https://github.com/openfaas/certifier.git
cd certifier
sha_cmd="sha256sum"
if [ ! -x "$(command -v $sha_cmd)" ]; then
    sha_cmd="shasum -a 256"
fi

if [ -x "$(command -v $sha_cmd)" ]; then
    sha_cmd="shasum"
fi

export SECRET=$(head -c 16 /dev/urandom| $sha_cmd | cut -d " " -f 1)

make test-kubernetes