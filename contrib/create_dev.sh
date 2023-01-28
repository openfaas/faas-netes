#!/usr/bin/env bash

# Creates an installation of OpenFaaS in kind for development

set -e
contrib/create_cluster.sh || exit 0;
contrib/deploy.sh
contrib/run_function.sh

PASSWORD=$(kubectl get secret -n openfaas basic-auth -o=go-template='{{index .data "basic-auth-password"}}' | base64 --decode)


cd "$(git rev-parse --show-toplevel)"
echo ""
echo ""
echo "Local dev cluster details:"
printf '%-10s:\t %s\n' 'Web UI' 'http://localhost:8080/ui'
printf '%-10s:\t %s\n' 'User' 'admin'
printf '%-10s:\t %s\n' 'Password' '$PASSWORD'
printf '%-10s:\t %s\n' 'CLI Login' "faas-cli login --username admin --password $PASSWORD"