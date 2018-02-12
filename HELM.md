# Helm

[Helm](https://github.com/kubernetes/helm) is a client CLI used to deploy Kubernetes
applications. It supports templating of configuration files, but also needs a server
component installing called `tiller`.

For more information on using Helm, refer to the Helm's [documentation](https://docs.helm.sh/using_helm/#quickstart-guide).

You can use this [chart](chart/openfaas) to install OpenFaaS to your cluster.

> Submission to the [main helm chart repository](https://github.com/kubernetes/charts) is pending.

## Pre-reqs:

* Instructions for latest Helm install

On Linux and Mac/Darwin:

```
$ curl https://raw.githubusercontent.com/kubernetes/helm/master/scripts/get | bash
```

Or via Homebrew on Mac:

```
$ brew install kubernetes-helm
```

* Create RBAC permissions for Tiller

```
$ kubectl -n kube-system create sa tiller \
 && kubectl create clusterrolebinding tiller \
      --clusterrole cluster-admin \
      --serviceaccount=kube-system:tiller
```

* Install the server-side Tiller component on your cluster

```
$ helm init --skip-refresh --upgrade --service-account tiller
```

> Note: this step installs a server component in your cluster. It can take anywhere between a few seconds to a few minutes to be installed properly. You should see tiller appear on: `kubectl get pods -n kube-system`.

## Deploy OpenFaaS via Helm

How to install chart please follow install instructions in chart's [readme](chart/openfaas/README.md).
