# Helm

[Helm](https://github.com/kubernetes/helm) is a client CLI used to deploy Kubernetes
applications. It supports templating of configuration files, but also needs a server
component installing called `tiller`.

## Pre-reqs:

Instructions for Kubernetes on Linux

* Install Helm

```
$ wget https://storage.googleapis.com/kubernetes-helm/helm-v2.6.1-linux-amd64.tar.gz
$ sudo tar -xvf helm-v2.6.1-linux-amd64.tar.gz --strip-components=1 -C /usr/local/bin/
```

* Create RBAC permissions for Tiller

```
kubectl -n kube-system create sa tiller \
 && kubectl create clusterrolebinding tiller \
  --clusterrole cluster-admin \
  --serviceaccount=kube-system:tiller
```

* Install the server-side Tiller component on your cluster

```
$ helm init --skip-refresh --upgrade --service-account tiller
```

## Deploy OpenFaaS via Helm

**Note:** You must also pass `--set rbac=false` if your cluster is not configured with role-based access control. For further information, see [here](https://kubernetes.io/docs/admin/authorization/rbac/).

---

To use defaults including the `default` Kubernetes namespace (recommended)

```
$ helm upgrade --install --debug --reset-values --set async=false openfaas openfaas/
```

For an advanced configuration with a separate function namespace:

```
$ kubectl create ns openfaas
$ kubectl create ns openfaas-fn

$ helm upgrade --install --debug --namespace openfaas \
  --reset-values --set async=false --set functionNamespace=openfaas-fn openfaas openfaas/
```

If you would like to enable asynchronous functions then use `--set async=true`. You can read more about asynchronous functions in the [OpenFaaS guides](https://github.com/openfaas/faas/tree/master/guide).

### OpenFaaS Helm chart options:

```
functionNamespace=defaults to the deployed namespace, kube namespace to create function deployments in
async=true/false defaults to false, deploy nats if true
armhf=true/false, defaults to false, use arm images if true (missing images for async)
exposeServices=true/false, defaults to true, will always create `ClusterIP` services, and expose `NodePorts` if true
rbac=true/false defaults to true, if true create roles
```


