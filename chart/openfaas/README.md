# OpenFaaS - Serverless Functions Made Simple

![OpenFaaS Logo](https://blog.alexellis.io/content/images/2017/08/faas_side.png)

[OpenFaaS](https://github.com/openfaas/faas) (Functions as a Service) is a framework for building serverless functions with Docker and Kubernetes which has first class support for metrics. Any process can be packaged as a function enabling you to consume a range of web events without repetitive boiler-plate coding.

**Highlights**

* Ease of use through UI portal and *one-click* install
* Write functions in any language for Linux or Windows and package in Docker/OCI image format
* Portable - runs on existing hardware or public/private cloud - [Kubernetes](https://github.com/openfaas/faas-netes) and Docker Swarm native
* [CLI](http://github.com/openfaas/faas-cli) available with YAML format for templating and defining functions
* Auto-scales as demand increases
* Scales to zero and back again
* Compatible with [Istio Service Mesh](https://istio.io). mTLS supported via `exec` health checks.

## Deploy OpenFaaS

**Note:** You must also pass `--set rbac=false` if your cluster is not configured with role-based access control. For further information, see [here](https://kubernetes.io/docs/admin/authorization/rbac/).

**Note:** If you can not use helm with Tiller, [skip below](#deployment-with-helm-template) for alternative install instructions.

---
### Install
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

Generate secrets so that we can enable basic authentication for the gateway:

```bash
# generate a random password
PASSWORD=$(head -c 12 /dev/urandom | shasum| cut -d' ' -f1)

kubectl -n openfaas create secret generic basic-auth \
--from-literal=basic-auth-user=admin \
--from-literal=basic-auth-password="$PASSWORD"
```

Now decide how you want to expose the services and edit the `helm upgrade` command as required.

* To use NodePorts (default) pass no additional flags
* To use a LoadBalancer add `--set serviceType=LoadBalancer`
* To use an IngressController add `--set ingress.enabled=true`

> Note: even without a LoadBalancer or IngressController you can access your gateway at any time via `kubectl port-forward`.

Now deploy OpenFaaS from the helm chart repo:

```
$ helm repo update \
 && helm upgrade openfaas --install openfaas/openfaas \
    --namespace openfaas  \
    --set basic_auth=true \
    --set functionNamespace=openfaas-fn
```

> The above command will also update your helm repo to pull in any new releases.

### Verify the installation
Once all the services are up and running, log into your gateway using the OpenFaaS CLI. This will cache your credentials into your `~/.openfaas/config.yml` file.

Fetch your public IP or NodePort via `kubectl get svc -n openfaas gateway-external -o wide` and set it as an environmental variable as below:

```bash
export OPENFAAS_URL=http://...
```

Now log in:

``` bash
echo -n $PASSWORD | faas-cli login -g $OPENFAAS_URL -u admin --password-stdin
```

## OpenFaaS Operator / CRD controller

If you would like to work with CRDs there is an alternative controller to faas-netes named [OpenFaaS Operator](https://github.com/openfaas-incubator/openfaas-operator) which can be swapped in at deployment time.
The OpenFaaS Operator is suitable for development and testing and may replace the faas-netes controller in the future.
The Operator is compatible with Kubernetes 1.9 or later.

To use it, add the flag: `--set operator.create=true` when installing with Helm.

### faas-netes vs OpenFaaS Operator

The faas-netes controller is the most tested, stable and supported version of the OpenFaaS integration with Kubernetes. In contrast the OpenFaaS Operator is based upon the codebase and features from `faas-netes`, but offers a tighter integration with Kubernetes through CustomResourceDefinitions. This means you can type in `kubectl get functions` for instance.

## Deployment with `helm template`
This option is good for those that have issues with installing Tiller, the server/cluster component of helm. Using the `helm` CLI, we can pre-render and then apply the templates using `kubectl`.

1. Clone the faas-netes repository
    ```sh
    $ git clone https://github.com/openfaas/faas-netes.git
    ```

2. Render the chart to a Kubernetes manifest called `openfaas.yaml`
    ```
    $ helm template faas-netes/chart/openfaas \
        --name openfaas \
        --namespace openfaas  \
        --set basic_auth=true \
        --set functionNamespace=openfaas-fn > $HOME/openfaas.yaml
    ```
    You can set the values and overrides just as you would in the install/upgrade commands above.

3. Install the components using `kubectl`
    ```
    $ kubectl apply -f faas-netes/namespaces.yml
    $ kubectl apply -f $HOME/openfaas.yaml
    ```

Now [verify your installation](#verify-the-installation).

## Deploy for development / testing

You can run the following command from within the `faas-netes/chart` folder in the `faas-netes` repo.

```
$ helm upgrade --install openfaas openfaas/ \
   --namespace openfaas \
   --set functionNamespace=openfaas-fn
```

## Exposing services

### NodePorts

By default a NodePort will be created for the API Gateway.

### LB

If you're running on a cloud such as AKS or GKE you will need to pass an additional flag of `--set serviceType=LoadBalancer` to tell `helm` to create LoadBalancer objects instead. An alternative to using multiple LoadBalancers is to install an Ingress controller.

### Deploy with an IngressController

In order to make use of automatic ingress settings you will need an IngressController in your cluster such as Traefik or Nginx.

Add `--set ingress.enabled` to enable ingress pass `--set ingress.enabled=true` when running the installation via `helm`.

By default services will be exposed with following hostnames (can be changed, see values.yaml for details):

* `gateway.openfaas.local`

### SSL / TLS

If you require TLS/SSL then please make use of an IngressController. A full guide is provided to [enable TLS for the OpenFaaS Gateway using cert-manager and Let's Encrypt](https://docs.openfaas.com/reference/ssl/kubernetes-with-cert-manager/).

## Zero scale

Scaling up from zero replicas is enabled by default, to turn it off set `zero_scale` to false in the helm chart.

Scaling to zero is done by the `faas-idler` component and by default will only carry out a dry-run. Pass the following to helm to enable scaling to zero replicas of idle functions. You will also need to [read the docs](https://docs.openfaas.com/architecture/autoscaling/#zero-scale) on how to configure functions to opt into scaling down.

```
--set faasIdler.dryRun=false
```

## Configuration

Additional OpenFaaS options in `values.yaml`.

| Parameter               | Description                           | Default                                                    |
| ----------------------- | ----------------------------------    | ---------------------------------------------------------- |
| `operator.create` | Use the OpenFaaS operator CRD controller, default uses faas-netes as the Kubernetes controller | `false` |
| `functionNamespace` | Functions namespace, preferred `openfaas-fn` | `default` |
| `async` | Deploys NATS | `true` |
| `exposeServices` | Expose `NodePorts/LoadBalancer`  | `true` |
| `serviceType` | Type of external service to use `NodePort/LoadBalancer` | `NodePort` |
| `ingress.enabled` | Create ingress resources | `false` |
| `rbac` | Enable RBAC | `true` |
| `basic_auth` | Enable basic authentication on the Gateway | `false` |
| `faasnetesd.readTimeout` | Queue worker read timeout | `60s` |
| `faasnetesd.writeTimeout` | Queue worker write timeout | `60s` |
| `faasnetesd.imagePullPolicy` | Image pull policy for deployed functions | `Always` |
| `gateway.replicas` | Replicas of the gateway, pick more than `1` for HA | `1` |
| `gateway.readTimeout` | Queue worker read timeout | `65s` |
| `gateway.writeTimeout` | Queue worker write timeout | `65s` |
| `gateway.upstreamTimeout` | Maximum duration of upstream function call, should be lower than `readTimeout`/`writeTimeout` | `60s` |
| `gateway.scaleFromZero` | Enables an intercepting proxy which will scale any function from 0 replicas to the desired amount | `true` |
| `gateway.maxIdleConns` | Set max idle connections from gateway to functions | `1024` |
| `gateway.maxIdleConnsPerHost` | Set max idle connections from gateway to functions per host | `1024` |
| `queueWorker.replicas` | Replicas of the queue-worker, pick more than `1` for HA | `1` |
| `queueWorker.ackWait` | Max duration of any async task/request | `60s` |
| `nats.enableMonitoring` | Enable the NATS monitoring endpoints on port `8222` | `false` |
| `openfaasImagePullPolicy` | Image pull policy for openfaas components, can change to `IfNotPresent` in offline env | `Always` |
| `kubernetesDNSDomain` | Domain name of the Kubernetes cluster | `cluster.local` |
| `faasIdler.inactivityDuration` | Duration after which faas-idler will scale function down to 0 | `5m` |
| `faasIdler.reconcileInterval` | The time between each of reconciliation | `30s` |
| `faasIdler.dryRun` | When set to false the OpenFaaS API will be called to scale down idle functions, by default this is set to only print in the logs. | `true` |

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
