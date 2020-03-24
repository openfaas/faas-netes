#!/bin/bash

set -e

# echo "go get -u sigs.k8s.io/kind@v0.4.0"

# export GO111MODULE="on"
# go get -u sigs.k8s.io/kind@v0.4.0

# Edited by Alex Ellis - KinD via binary is *much* faster, 
# Go modules now adds several minutes to the build / test time :-(

export VER="v0.7.0"
echo "Downloading KinD $VER"
curl -SLfs https://github.com/kubernetes-sigs/kind/releases/download/$VER/kind-linux-amd64 > kind-linux-amd64
chmod +x kind-linux-amd64 
sudo mv kind-linux-amd64 /usr/local/bin/kind