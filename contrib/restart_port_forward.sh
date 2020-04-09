#!/usr/bin/env bash

DEVENV=${OF_DEV_ENV:-kind}

if [ -f "of_${DEVENV}_portforward.pid" ]; then
    kill $(<of_${DEVENV}_portforward.pid)
fi

# quietly start portforward and put it in the background, it will not
# print every connection handled
kubectl --context "kind-$DEVENV" port-forward deploy/gateway -n openfaas 31112:8080 &>/dev/null & \
    echo -n "$!" > "of_${DEVENV}_portforward.pid"
