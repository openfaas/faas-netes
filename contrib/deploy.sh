#!/usr/bin/env bash

set -e

DEVENV=${OF_DEV_ENV:-kind}
OPERATOR=${OPERATOR:-0}

echo "Applying namespaces"
kubectl --context "kind-$DEVENV" apply -f ./namespaces.yml

sha_cmd="sha256sum"
if [ ! -x "$(command -v $sha_cmd)" ]; then
    sha_cmd="shasum -a 256"
fi

if [ -x "$(command -v $sha_cmd)" ]; then
    sha_cmd="shasum"
fi

PASSWORD=$(head -c 16 /dev/urandom| $sha_cmd | cut -d " " -f 1)
echo -n $PASSWORD > password.txt

kubectl --context "kind-$DEVENV" -n openfaas create secret generic basic-auth \
--from-literal=basic-auth-user=admin \
--from-literal=basic-auth-password="$PASSWORD"

CREATE_OPERATOR=false
if [ "${OPERATOR}" == "1" ]; then
    CREATE_OPERATOR="true"
fi

echo "Waiting for helm install to complete."

helm upgrade \
    --kube-context "kind-$DEVENV" \
    --install \
    openfaas \
    ./chart/openfaas \
    --namespace openfaas  \
    --set basic_auth=true \
    --set functionNamespace=openfaas-fn \
    --set operator.create=$CREATE_OPERATOR

if [ "${OPERATOR}" == "1" ]; then

    kubectl --context "kind-$DEVENV" patch -n openfaas deploy/gateway \
      -p='[{"op": "add", "path": "/spec/template/spec/containers/1/command", "value": ["./faas-netes", "-operator=true"]} ]' --type=json
fi

kubectl --context "kind-$DEVENV" rollout status deploy/prometheus -n openfaas
kubectl --context "kind-$DEVENV" rollout status deploy/gateway -n openfaas
