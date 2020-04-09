#!/usr/bin/env bash

OPERATOR=${OPERATOR:-0}

CREATE_OPERATOR=false

if [ "${OPERATOR}" == "1" ]; then
    CREATE_OPERATOR="true"
fi

echo "Waiting for helm install to complete."

echo helm upgrade \
    --install \
    openfaas \
    ./chart/openfaas \
    --namespace openfaas  \
    --set basic_auth=true \
    --set functionNamespace=openfaas-fn \
    --set operator.create=$CREATE_OPERATOR \
    --wait
