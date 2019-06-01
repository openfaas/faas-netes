#!/usr/bin/env bash

DEVENV=${OF_DEV_ENV:-kind}

export KUBECONFIG="$(kind get kubeconfig-path --name="$DEVENV")"

kubectl rollout status deploy/gateway -n openfaas --timeout=1m

if [ $? != 0 ];
then
   exit 1
fi


if [ -f "of_${DEVENV}_portforward.pid" ]; then
    kill $(<of_${DEVENV}_portforward.pid)
fi

# quietly start portforward and put it in the background, it will not
# print every connection handled
kubectl port-forward deploy/gateway -n openfaas 31112:8080 &>/dev/null & \
    echo -n "$!" > "of_${DEVENV}_portforward.pid"

# port-forward needs some time to start
sleep 10

# Login in OpenFaas
export OPENFAAS_URL=http://127.0.0.1:31112

cat ./password.txt | faas-cli login --username admin --password-stdin

faas-cli deploy --image=functions/alpine:latest --fprocess=cat --name "echo"

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
