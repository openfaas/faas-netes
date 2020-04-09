#!/usr/bin/env bash

DEVENV=${OF_DEV_ENV:-kind}

if [ -f "of_${DEVENV}_portforward.pid" ]; then
    kill $(<of_${DEVENV}_portforward.pid)
    rm "of_${DEVENV:-kind}_portforward.pid"
fi

kind delete cluster --name "$DEVENV"
