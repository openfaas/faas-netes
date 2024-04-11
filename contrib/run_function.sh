#!/usr/bin/env bash

set -e

DEVENV=${OF_DEV_ENV:-kind}
OPERATOR=${OPERATOR:-0}

kubectl --context "kind-$DEVENV" rollout status deploy/gateway -n openfaas --timeout=1m

if [ $? != 0 ];
then
   exit 1
fi

if [ -f "of_${DEVENV}_portforward.pid" ]; then
    kill "$(<of_${DEVENV}_portforward.pid)" > /dev/null 2> /dev/null || :
fi

# quietly start portforward and put it in the background, it will not
# print every connection handled
kubectl --context "kind-$DEVENV" port-forward deploy/gateway -n openfaas 8080:8080 &>/dev/null & \
    echo -n "$!" > "of_${DEVENV}_portforward.pid"

# Wait for the gateway to be ready
faas-cli ready --attempts=180 --interval=1s

echo "Using existing password for OpenFaaS gateway retrieved from Kubernetes."
PASSWORD=$(kubectl get secret -n openfaas basic-auth -o=go-template='{{index .data "basic-auth-password"}}' | base64 --decode)
echo -n $PASSWORD > password.txt

export OPENFAAS_URL=http://127.0.0.1:8080

# Login into the gateway
echo ""
echo "Login to the gateway..."
cat ./password.txt | faas-cli login --username admin --password-stdin

IS_DEPLOYED=$(faas-cli list | grep echo | wc -c)
if [ $IS_DEPLOYED -gt 0 ];then
    echo ""
    echo "test function 'echo' already deployed"
    exit;
fi

# Deploy via the REST API which can test either the controller or operator
faas-cli deploy --image=functions/alpine:latest --fprocess=cat --name "echo"

# Call echo function
faas-cli ready --attempts=180 --interval=1s echo
