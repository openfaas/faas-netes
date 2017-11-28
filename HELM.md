# Helm

[Helm](https://github.com/kubernetes/helm) is a client CLI used to deploy Kubernetes
applications. It supports templating of configuration files, but also needs a server
component installing called `tiller`.

You can use this chart to install OpenFaaS to your cluster.

> Submission to the [main helm chart repository](https://github.com/kubernetes/charts) is pending.

## Pre-reqs:

Instructions for Kubernetes on Linux

* Install Helm

```
$ wget https://storage.googleapis.com/kubernetes-helm/helm-v2.6.2-linux-amd64.tar.gz && \
  sudo tar -xvf helm-v2.6.2-linux-amd64.tar.gz --strip-components=1 -C /usr/local/bin/
```

On Mac/Darwin:

```
$ wget https://storage.googleapis.com/kubernetes-helm/helm-v2.6.2-darwin-amd64.tar.gz && \
  sudo tar -xvf helm-v2.6.2-darwin-amd64.tar.gz --strip-components=1 -C /usr/local/bin/

```

Or via Homebrew on Mac:

```
$ brew install kubernetes-helm
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

Clone the faas-netes repo including the helm chart:

```
$ git clone https://github.com/openfaas/faas-netes && \
  cd faas-netes
```

To use defaults including the `default` Kubernetes namespace (recommended)

```
$ helm upgrade --install --debug --reset-values --set async=false openfaas openfaas/
```

Optional: you can also a separate namespace for functions:

```
$ kubectl create ns openfaas
$ kubectl create ns openfaas-fn

$ helm upgrade --install --debug --namespace openfaas \
  --reset-values --set async=false --set functionNamespace=openfaas-fn openfaas openfaas/
```

If you would like to enable asynchronous functions then use `--set async=true`. You can read more about asynchronous functions in the [OpenFaaS guides](https://github.com/openfaas/faas/tree/master/guide).

By default you will have NodePorts available for each service such as the API gateway and Prometheus

### Deploy with an IngressController

In order to make use of automatic ingress settings you will need an IngressController in your cluster such as Traefik or Nginx.

Add `--set ingress.enabled` to enable ingress:

```
$ helm upgrade --install --debug --reset-values --set async=false --set ingress.enabled=true openfaas openfaas/
```

By default services will be exposed with following hostnames (can be changed, see values.yaml for details):
* `faas-netesd.openfaas.local`
* `gateway.openfaas.local`
* `prometheus.openfaas.local`
* `alertmanager.openfaas.local`

### Additional OpenFaaS Helm chart options:

* `functionNamespace=defaults` - to the deployed namespace, kube namespace to create function deployments in
* `async=true/false` - defaults to false, deploy nats if true
* `armhf=true/false` - defaults to false, use arm images if true (missing images for async)
* `exposeServices=true/false` - defaults to true, will always create `ClusterIP` services, and expose `NodePorts/LoadBalancer` if true (based on serviceType)
* `serviceType=NodePort/LoadBalancer` - defaults to NodePort, type of external service to use when exposeServices is set to true
* `rbac=true/false` - defaults to true, if true create roles
* `ingress.enabled=true/false` - defaults to false, set to true to create ingress resources. See openfaas/values.yaml for detailed Ingress configuration.

### Removing the OpenFaaS Helm chart

All control plane components can be cleaned up with helm with:

```bash
helm delete --purge openfaas
```

Individual functions will need to be either deleted before deleting the chart with `faas-cli` or manually deleted using `kubectl delete`.
