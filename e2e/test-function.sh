#!/usr/bin/env bash

set -o errexit

REPO_ROOT=$(git rev-parse --show-toplevel)
TEMP_DIR=$(mktemp -d)

logs() {
  kubectl -n openfaas logs deployment/gateway -c operator
}
trap "logs" EXIT SIGINT

echo ">>> Create test secrets"
kubectl -n openfaas-fn create secret generic faas-token --from-literal=faas-token=token
kubectl -n openfaas-fn create secret generic faas-key --from-literal=faas-key=key

echo ">>> Create test function"
kubectl apply -f ${REPO_ROOT}/artifacts/nodeinfo.yaml

echo '>>> Waiting for function deployment'
retries=10
count=0
ok=false
until ${ok}; do
    kubectl -n openfaas-fn get deployment/nodeinfo | grep 'nodeinfo' && ok=true || ok=false
    sleep 5
    count=$(($count + 1))
    if [[ ${count} -eq ${retries} ]]; then
        kubectl -n openfaas-fn describe pods
        echo "No more retries left"
        exit 1
    fi
done

echo '>>> Waiting for function deployment to be ready'
kubectl -n openfaas-fn rollout status deployment/nodeinfo

echo ">>> Starting port forwarding"
kubectl -n openfaas port-forward svc/gateway 31112:8080 &>/dev/null & \
echo -n "$!" > "${TEMP_DIR}/portforward.pid"

# wait for port forwarding to start
sleep 10

echo '>>> Invoke function'
retries=10
count=0
ok=false
until ${ok}; do
    curl -sd '' -X POST http://127.0.0.1:31112/function/nodeinfo | grep 'Hostname' && ok=true || ok=false
    sleep 5
    count=$(($count + 1))
    if [[ ${count} -eq ${retries} ]]; then
        kubectl -n openfaas-fn describe pods
        echo "No more retries left"
        exit 1
    fi
done

