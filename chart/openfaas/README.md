# OpenFaaS - Serverless Functions Made Simple

![OpenFaaS Logo](https://blog.alexellis.io/content/images/2017/08/faas_side.png)

[OpenFaaS](https://github.com/openfaas/faas) (Functions as a Service) is a framework for building serverless functions with Docker and Kubernetes which has first class support for metrics. Any process can be packaged as a function enabling you to consume a range of web events without repetitive boiler-plate coding.

**Highlights**

* Ease of use through UI portal and *one-click* install
* Write functions in any language for Linux or Windows and package in Docker/OCI image format
* Portable - runs on existing hardware or public/private cloud - [Kubernetes](https://github.com/openfaas/faas-netes) and Docker Swarm native
* [CLI](http://github.com/openfaas/faas-cli) available with YAML format for templating and defining functions
* Auto-scales as demand increases

## Deploy OpenFaaS

**Note:** You must also pass `--set rbac=false` if your cluster is not configured with role-based access control. For further information, see [here](https://kubernetes.io/docs/admin/authorization/rbac/).

---

We recommend creating two namespaces, one for the OpenFaaS core services and one for the functions:

```
$ kubectl apply -f https://raw.githubusercontent.com/openfaas/faas-netes/master/namespaces.yml
```

You will now have `openfaas` and `openfaas-fn`. If you want to change the names or to install into multiple installations then edit `namespaces.yml` from the `faas-netes` repo.

Add the OpenFaaS `helm` chart:

```bash
$ helm repo add openfaas https://openfaas.github.io/faas-netes/
"openfaas" has been added to your repositories
```

Now deploy OpenFaaS from the helm chart repo:

```
$ helm repo update \
 && helm upgrade openfaas --install openfaas/openfaas \
    --namespace openfaas  \
    --set functionNamespace=openfaas-fn
```

The above command will also update your helm repo to pull in any new releases.

## OpenFaaS Operator / CRD controller

If you would like to work with CRDs there is an alternative controller to faas-netes named [OpenFaaS Operator](https://github.com/openfaas-incubator/openfaas-operator) which can be swapped in at deployment time.
The OpenFaaS Operator is suitable for development and testing and may replace the faas-netes controller in the future.
The Operator is compatible with Kubernetes 1.9 or later.

To use it, add the flag: `--set operator.create=true` when installing with Helm.

### faas-netes vs OpenFaaS Operator

The faas-netes controller is the most tested, stable and supported version of the OpenFaaS integration with Kubernetes. In contrast the OpenFaaS Operator is based upon the codebase and features from `faas-netes`, but offers a tighter integration with Kubernetes through 
CustomResourceDefinitions. This means you can type in `kubectl get functions` for instance.

## Deploy for development / testing

You can run the following command from within the `faas-netes/chart` folder in the `faas-netes` repo.

```
$ helm upgrade --install openfaas openfaas/ \
   --namespace openfaas \
   --set functionNamespace=openfaas-fn
```

## Exposing services

### NodePorts

By default NodePorts will be created for the API Gateway and Prometheus.

### LB

If you're running on a cloud such as AKS or GKE you will need to pass an additional flag of `--set serviceType=LoadBalancer` to tell `helm` to create LoadBalancer objects instead. An alternative to using multiple LoadBalancers is to install an Ingress controller.

### Deploy with an IngressController

In order to make use of automatic ingress settings you will need an IngressController in your cluster such as Traefik or Nginx.

Add `--set ingress.enabled` to enable ingress pass `--set ingress.enabled=true` when running the installation via `helm`.

By default services will be exposed with following hostnames (can be changed, see values.yaml for details):

* `faas-netes.openfaas.local`
* `gateway.openfaas.local`
* `prometheus.openfaas.local`
* `alertmanager.openfaas.local`

### Secure the Gateway administrative API and UI with basic auth

In order to enable basic auth first you need to create a secret named basic-auth in the openfaas namespace:

```bash
# generate a random password
password=$(head -c 12 /dev/urandom | shasum| cut -d' ' -f1)

kubectl -n openfaas create secret generic basic-auth \
--from-literal=basic-auth-user=admin \
--from-literal=basic-auth-password=$password 
```

Update the Helm release and enable basic auth with:

```
helm upgrade --reuse-values openfaas openfaas/openfaas \
    --set basic_auth=true
```

Save your credentials in faas-cli store:

```bash
echo $password | faas-cli login -g http://GATEWAY-URL -u admin --password-stdin
```

## Configuration

Additional OpenFaaS options in `values.yaml`.

| Parameter               | Description                           | Default                                                    |
| ----------------------- | ----------------------------------    | ---------------------------------------------------------- |
| `operator.create` | Use the OpenFaaS operator CRD controller, default uses faas-netes as the Kubernetes controller | `false` |
| `functionNamespace` | Functions namespace | `default` |
| `async` | Deploys NATS | `true` |
| `exposeServices` | Expose `NodePorts/LoadBalancer`  | `true` |
| `serviceType` | Type of external service to use `NodePort/LoadBalancer` | `NodePort` |
| `ingress.enabled` | Create ingress resources | `false` |
| `rbac` | Enable RBAC | `true` |
| `faasnetesd.readTimeout` | Queue worker read timeout | `20s` |
| `faasnetesd.writeTimeout` | Queue worker write timeout | `20s` |
| `faasnetesd.imagePullPolicy` | Image pull policy for deployed functions | `Always` |
| `gateway.readTimeout` | Queue worker read timeout | `20s` |
| `gateway.writeTimeout` | Queue worker write timeout | `20s` |
| `gateway.upstreamTimeout` | Maximum duration of upstream function call | `20s` |
| `queueWorker.ackWait` | Max duration of any async task/request | `30s` |
| `openfaasImagePullPolicy` | Image pull policy for openfaas components, can change to `IfNotPresent` in offline env | `Always` |

Specify each parameter using the `--set key=value[,key=value]` argument to `helm install`.
See values.yaml for detailed configuration.

## Removing the OpenFaaS

All control plane components can be cleaned up with helm:

```
$ helm delete --purge openfaas
```

Follow this by the following to remove all other associated objects:

```
$ kubectl delete namespace/openfaas
$ kubectl delete namespace/openfaas-fn
```

In some cases your additional functions may need to be either deleted before deleting the chart with `faas-cli` or manually deleted using `kubectl delete`.
