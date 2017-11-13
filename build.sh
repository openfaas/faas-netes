#!/bin/sh

make build

# os="$(uname -s)"

# if [ "$os" = "Linux" ] ; then
# 	cd openfaas
# 	if ! [ -x "$(command -v helm)" ]; then
# 		curl -SL https://storage.googleapis.com/kubernetes-helm/helm-v2.6.1-linux-amd64.tar.gz > /tmp/helm.tgz && \
# 		mkdir -p /tmp/helm/ && \
#                 tar -xvf helm.tgz --strip-components=1 -C /tmp/helm
# 		/tmp/helm/helm lint
# 	else
# 		helm lint	
# 	fi
# else
# 	echo "Only checking helm lint on Linux"
# fi

