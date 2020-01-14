# Helm

[Helm](https://github.com/kubernetes/helm) is a client CLI used to deploy Kubernetes
applications.

## Helm 3

Get the binary:

```sh
curl -sSLf https://raw.githubusercontent.com/helm/helm/master/scripts/get-helm-3 | bash
```

That's it, no tiller is required for using helm3.

Now deploy OpenFaaS via Helm: using the [Chart's readme](chart/openfaas/README.md).

## Helm 2 (legacy)

It supports templating of configuration files, but also needs a server
component installing called `tiller`.

For more information on using Helm, refer to the Helm's [documentation](https://docs.helm.sh/using_helm/#quickstart-guide).

You can use this [chart](chart/openfaas) to install OpenFaaS to your cluster.

> Submission to the [main helm chart repository](https://github.com/kubernetes/charts) is pending.

## Pre-reqs:

### Install the helm CLI/client

Instructions for latest Helm install

* On Linux and Mac/Darwin:

      curl https://raw.githubusercontent.com/kubernetes/helm/master/scripts/get | bash

* Or via Homebrew on Mac:

      brew install kubernetes-helm

### Install tiller

* Create RBAC permissions for tiller

```sh
kubectl -n kube-system create sa tiller \
  && kubectl create clusterrolebinding tiller \
  --clusterrole cluster-admin \
  --serviceaccount=kube-system:tiller
```

* Install the server-side Tiller component on your cluster

```sh
helm init --skip-refresh --upgrade --service-account tiller
```

> Note: this step installs a server component in your cluster. It can take anywhere between a few seconds to a few minutes to be installed properly. You should see tiller appear on: `kubectl get pods -n kube-system`.

Now deploy OpenFaaS via Helm: using the [Chart's readme](chart/openfaas/README.md).