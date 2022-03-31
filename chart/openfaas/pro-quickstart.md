## OpenFaaS Pro quick-start for local development

```bash
kind create cluster --name pro
```

```bash
git clone https://github.com/openfaas/faas-netes --depth=1
cd faas-netes

kubectl apply -f namespaces.yml

kubectl create secret generic \
    -n openfaas \
    openfaas-license \
    --from-file license=$HOME/.openfaas/LICENSE

helm upgrade openfaas --install chart/openfaas \
    --namespace openfaas \
    --set functionNamespace=openfaas-fn \
    --set generateBasicAuth=true \
    -f ./chart/openfaas/values.yaml \
    -f ./chart/openfaas/values-pro.yaml \
    --set gateway.replicas=1 \
    --set queueWorker.replicas=1
```

Or install with dashboard:

```bash
helm upgrade openfaas --install chart/openfaas \
    --namespace openfaas \
    --set functionNamespace=openfaas-fn \
    --set generateBasicAuth=true \
    -f ./chart/openfaas/values.yaml \
    -f ./chart/openfaas/values-pro.yaml \
    --set gateway.replicas=1 \
    --set queueWorker.replicas=1 \
    --set dashboard.enabled=true \
    --set dashboard.publicURL=https://portal.exit.o6s.io
```

Confirm faas-netes and queue-worker versions are from ghcr.io/openfaasltd

Confirm that autoscaler is deployed.

```bash
kubectl -n openfaas get deployments -o wide
```

Confirm running in operator mode:

```bash
kubectl logs -n openfaas deploy/gateway -c operator
```

