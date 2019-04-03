#!/bin/bash

export KUBECONFIG="$(kind get kubeconfig-path)"

kubectl rollout status deploy/gateway -n openfaas --timeout=1m

if [ $? != 0 ];
then
   exit 1
fi

kubectl port-forward deploy/gateway -n openfaas 8080:8080 &

# port-forward needs some time to start
sleep 10

# Login in OpenFaas
export OPENFAAS_URL=http://127.0.0.1:8080
PASSWORD=$(cat ./password.txt)

faas-cli login --username admin --password $PASSWORD

faas-cli deploy --image=functions/alpine:latest --fprocess=cat --name echo

# Call echo function
for i in {1..180};
do
    Ready="$(faas-cli describe echo | awk '{ if($1 ~ /Status:/) print $2 }')"
    if [[ $Ready == "Ready" ]];
    then
        exit 0
    fi
    sleep 1
done

exit 1
