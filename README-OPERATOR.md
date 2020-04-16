faas-netes
===========

faas-netes operator README

### Example `Function` for `openfaas.com/v1` CRD

* Minimal example:

```yaml
apiVersion: openfaas.com/v1
kind: Function
metadata:
  name: nodeinfo
  namespace: openfaas-fn
spec:
  name: nodeinfo
  image: functions/nodeinfo:latest
```

* Extended example:

```yaml
apiVersion: openfaas.com/v1
kind: Function
metadata:
  name: nodeinfo
  namespace: openfaas-fn
spec:
  name: nodeinfo
  handler: node main.js
  image: functions/nodeinfo:latest
  labels:
    com.openfaas.scale.min: "2"
    com.openfaas.scale.max: "15"
  annotations:
    current-time: Mon 6 Aug 23:42:00 BST 2018
    next-time: Mon 6 Aug 23:42:00 BST 2019
  environment:
    write_debug: "true"
  limits:
    cpu: "200m"
    memory: "256Mi"
  requests:
    cpu: "10m"
    memory: "128Mi"
  constraints:
    - "cloud.google.com/gke-nodepool=default-pool"
  secrets:
    - nodeinfo-secret1
```

*Example adapted from [artifacts/nodeinfo.yaml](./artifacts/nodeinfo.yaml)*

* Generate CRD from `stack.yml`

You can generate CRD entries separated by `---` by running `faas-cli generate` in the same folder as a `stack.yml` file. Each function in the file will be outputted. If you want a CRD for one function only then you can also pass `--filter=function-name`.

```bash
# create a go function for Docker Hub user `alexellis2`
faas-cli new --lang go --prefix alexellis2 crd-example

# build and push an image
faas-cli build -f crd-example.yaml
faas-cli push -f crd-example.yaml

# generate the CRD entry from the "stack.yml" file and apply in the cluster
faas-cli generate -f crd-example.yaml | kubectl apply -f -
```

* Generate CRD from Function Store

```bash

# find a function in the store
faas-cli store list
...


# generate to a file
faas-cli generate --from-store="figlet" > figlet-crd.yaml
kubectl apply -f figlet-crd.yaml
```

## Get started

You can deploy OpenFaaS and the operator using helm, or generate static YAML through the `helm template` command.

### Deploy OpenFaaS with the operator

Two namespaces will be used - `openfaas` for the core services and `openfaas-fn` for functions.

```bash
# create OpenFaaS namespaces
kubectl apply -f https://raw.githubusercontent.com/openfaas/faas-netes/master/namespaces.yml

# add OpenFaaS Helm repo
helm repo add openfaas https://openfaas.github.io/faas-netes/

# generate a random password
PASSWORD=$(head -c 12 /dev/urandom | shasum| cut -d' ' -f1)

kubectl -n openfaas create secret generic basic-auth \
--from-literal=basic-auth-user=admin \
--from-literal=basic-auth-password="$PASSWORD"

# get latest chart version
helm repo update

# install with basic auth enabled
helm upgrade openfaas --install openfaas/openfaas \
    --namespace openfaas  \
    --set basic_auth=true \
    --set functionNamespace=openfaas-fn \
    --set operator.create=true
```

> Note: If you are switching from the OpenFaaS `faas-netes` controller, then you will need to remove all functions and redeploy them after switching to the operator.

#### Deploy a function with kubectl:

```bash
kubectl -n openfaas-fn apply -f ./artifacts/nodeinfo.yaml
```

#### List functions, services, deployments and pods

```bash
kubectl -n openfaas-fn get functions
kubectl -n openfaas-fn get all
``` 

#### Deploy a function with secrets

```bash
kubectl -n openfaas-fn create secret generic faas-token --from-literal=faas-token=token
kubectl -n openfaas-fn create secret generic faas-key --from-literal=faas-key=key
```

Add the secrets section in `nodeinfo.yaml` and re-apply the function:

```yaml
  secrets:
   - faas-token
   - faas-key
```

Test that secrets are available inside the `nodeinfo` pod:

```bash
kubectl -n openfaas-fn exec -it nodeinfo-84fd464784-sd5ml -- sh

~ $ cat /var/openfaas/faas-key 
key

~ $ cat /var/openfaas/faas-token 
token
``` 

Test that node selectors work on GKE by adding the following to `nodeinfo.yaml`:

```yaml
  constraints:
    - "cloud.google.com/gke-nodepool=default-pool"
```

Apply the function and check the deployment specs with:

```bash
kubectl -n openfaas-fn describe deployment nodeinfo
```

#### Development build

The OpenFaaS Operator runs as a sidecar in the gateway pod. For end to end testing you need to update the sidecar to use
your development image.

1. Build, tag and push your image to your own public docker repository: 

```bash
export USERNAME="username"

make build \
 && docker tag openfaas/openfaas-operator:latest $USERNAME/openfaas-operator:latest-dev \
 && docker push $USERNAME/openfaas-operator:latest-dev
```

2. Update your helm deployment

```bash
helm upgrade openfaas --install openfaas/openfaas \
    --namespace openfaas \
    --set functionNamespace=openfaas-fn \
    --set operator.create=true \
--set operator.image=$USERNAME/openfaas-operator:latest-dev
```

### Run locally, without Docker

You will a KUBECONFIG file and accessible in your path.

* Create the OpenFaaS CRD

```bash
$ kubectl apply -f artifacts/operator-crd.yaml
```

* Build and launch the OpenFaaS controller locally

```bash
$ go build \
  && ./openfaas-operator -kubeconfig=$HOME/.kube/config -v=4
```

As an alternative, you can use `go run`

```bash
$ go run *.go -kubeconfig=$HOME/.kube/config
```

To use an alternative port set the `port` environmental variable to another value.

Create a function:

```bash
$ kubectl apply -f artifacts/nodeinfo.yaml
```

Check if nodeinfo deployment and service were created through the CRD:

```bash
$ kubectl get deployment nodeinfo
$ kubectl get service nodeinfo
```

Test if nodeinfo service can access the pods:

```bash
$ kubectl run -it --rm --restart=Never curl --image=byrnedo/alpine-curl:latest --command -- sh
/ # curl -d 'verbose' http://nodeinfo.default:8080
```

Delete nodeinfo function:

```bash
kubectl delete -f artifacts/nodeinfo.yaml 
```

Check if nodeinfo pods, rc, deployment and service were removed:
```bash
kubectl get all
```

### REST API

The operator also implements the OpenFaaS REST API which provides an additional way to manage functions and secrets in addition to using the CRD with `kubectl` directly.

If OpenFaaS is configured with the `basic_auth=true` flag then Basic Authentication is enabled on the REST API. If that is the case then reformat each `curl` command to also include the credentials.

* Basic Auth off

```bash
curl -s http://localhost:8081/system/functions | jq .
```

* Basic Auth enabled

```bash
export USER=""
export PASSWORD=""

curl -s http://$USER:$PASSWORD@localhost:8081/system/functions | jq .
```

#### Function management

Create or update a function:

```bash
curl -d '{"service":"nodeinfo","image":"functions/nodeinfo:burner","envProcess":"node main.js","labels":{"com.openfaas.scale.min":"2","com.openfaas.scale.max":"15"},"environment":{"output":"verbose","debug":"true"}}' -X POST  http://localhost:8081/system/functions
```

List functions:

```bash
curl -s http://localhost:8081/system/functions | jq .
```

Scale PODs up/down:

```bash
curl -d '{"serviceName":"nodeinfo", "replicas": 1}' -X POST http://localhost:8081/system/scale-function/nodeinfo
```

Get available replicas:

```bash
curl -s http://localhost:8081/system/function/nodeinfo | jq .availableReplicas
```

Remove function:

```bash
curl -d '{"functionName":"nodeinfo"}' -X DELETE http://localhost:8081/system/functions
```

#### Secret management

Create secret:

```bash
curl -d '{"name":"test","value":"test"}' -X POST http://localhost:8081/system/secrets
```

List secrets:

```bash
curl -X GET http://localhost:8081/system/secrets
```

Update secret:

```bash
curl -d '{"name":"test","value":"test update"}' -X PUT http://localhost:8081/system/secrets
```

Delete secret:

```bash
curl -d '{"name":"test"}' -X DELETE http://localhost:8081/system/secrets
```

#### Configure a service account for your function

Example service account:

```yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  name: build-robot
  namespace: openfaas-fn
```

Set the following in your function spec:

```yaml
annotations:
  com.openfaas.serviceaccount: "build-robot"
```

### Logging

Verbosity levels:

* `-v=0` CRUD actions via API and Controller including errors
* `-v=2` function call duration (Proxy API)
* `-v=4` Kubernetes informers events (highly verbose)

### Instrumentation

Prometheus route:

```bash
curl http://localhost:8081/metrics
```

Profiling is disabled by default, to enable it set `pprof` environment variable to `true`.

The `pprof` UI can be access at `http://localhost:8081/debug/pprof/`. The goroutine, heap and threadcreate 
profilers are enabled along with the full goroutine stack dump.

Run the heap profiler:

```bash
go tool pprof goprofex http://localhost:8081/debug/pprof/heap
```

Run the goroutine profiler:

```bash
go tool pprof goprofex http://localhost:8081/debug/pprof/goroutine
```
