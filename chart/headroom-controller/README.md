# Headroom Controller Helm Chart

Speed up scaling up and rolling out of workloads by pre-provisioning headroom in your Kubernetes cluster.

This component is licensed for OpenFaaS Standard or Enterprise customers.

## Prerequisites

You do not need to use OpenFaaS to install this chart and use the headroom-controller. It is a standalone component.

- Kubernetes
- Helm
- Some form of autoscaling node group, or Cluster Autoscaler, or Karpenter, etc.
- An OpenFaaS license

  Obtain a license or trial at https://openfaas.com/pricing

If you're not an OpenFaaS Standard or OpenFaaS for Enterprises customer, you can [purchase a license here](https://subscribe.openfaas.com/). Click the "Headroom Controller" option.

A license is required for each cluster/installation.

You can build a direct checkout link as follows, changing `?quantity=N` for instance:

```
# 2x clusters:
https://subscribe.openfaas.com/checkout/buy/6f00c083-378f-41dd-ab01-aaf2ebfac73f?quantity=2


# 3x clusters:
https://subscribe.openfaas.com/checkout/buy/6f00c083-378f-41dd-ab01-aaf2ebfac73f?quantity=2
```

## Demo

Watch a [live demo with the Cluster Autoscaler and K3s](https://www.youtube.com/embed/MHXvhKb6PpA?si=QFEf632Ha3VUESbs):

[![Video demo of headroom controller](https://img.youtube.com/vi/MHXvhKb6PpA/0.jpg)](https://www.youtube.com/watch?v=MHXvhKb6PpA)

## Installation

### Add the repository

```bash
helm repo add openfaas https://openfaas.github.io/faas-netes/ && \
  helm repo update
```

### Install from the chart repository

Create the required secret with your license:

```bash
kubectl create ns default

kubectl create secret generic \
  -n default \
  openfaas-license \
  --from-file license=$HOME/.openfaas/LICENSE
```

> OpenFaaS customers - use your existing license key for the specific environment. If you're not an OpenFaaS customer, use the key you received from LemonSqueezy via email.

Install with defaults:

```bash
helm repo update

helm upgrade --install headroom-controller \
  openfaas/headroom-controller \
  --namespace default
```

Or create a file named `values-custom.yaml` if you wish to customise the installation.

```yaml
rbac:
  role: Role
```

Then install with custom values:

```bash
helm repo update

helm upgrade --install headroom-controller \
  openfaas/headroom-controller \
  --namespace default \
  --values values-custom.yaml
```


The chart can also be installed into a custom namespace i.e. `openfaas` or `kube-system` by altering the `--namespace` flag.

If you install into a different namespace, create the `openfaas-license` secret in that namespace first.

### Install from local chart

If you are working on a patch for the helm chart, you can deploy it directly from a local folder.

```bash
# Clone the repository
git clone https://github.com/openfaas/faas-netes.git --depth=1
cd faas-netes

# Install the chart
helm upgrade --install headroom-controller \
  ./chart/headroom-controller \
  --namespace default
```

## Usage

1. Create a default priority priorityClassName

```bash
kubectl apply -f - << EOF
apiVersion: scheduling.k8s.io/v1
kind: PriorityClass
metadata:
  name: default
value: 1000
globalDefault: true
description: "Default priority class for all pods"
EOF
```

2. Create a low priority class for the headroom Custom Resources

```bash
kubectl apply -f - <<EOF
apiVersion: scheduling.k8s.io/v1
kind: PriorityClass
metadata:
  name: headroom
description: Low priority class for headroom pods
globalDefault: false
preemptionPolicy: Never
value: -10
EOF
```

3. Create a headroom resource:

```bash
kubectl apply -f - <<EOF
apiVersion: openfaas.com/v1
kind: Headroom
metadata:
  name: example
spec:
  replicas: 1
  requests:
    cpu: "100m"
    memory: "128Mi"
  priorityClassName: "headroom"
EOF
```

4. Check the status:

   `kubectl get headrooms`

5. Scale the headroom:

   `kubectl scale headroom example --replicas=2`

## Advanced example - scale a headroom down from a Cron Job

Imagine you wanted to scale down the headroom resource overnight.

Here's an example of how Kubernetes-native resources can be used.

```yaml
kind: CronJob
apiVersion: batch/v1
metadata:
  name: scale-headroom
  namespace: default
spec:
  schedule: "0 0 * * * *"
  successfulJobsHistoryLimit: 5 # Keep 5 successful jobs
  failedJobsHistoryLimit: 1 # Keep 1 failed job
  jobTemplate:
    spec:
      template:
        spec:
          serviceAccountName: headroom-scaler
          restartPolicy: OnFailure
          containers:
            - name: kubectl
              image: alpine/k8s:1.33.3 # Or a specific version
              command:
                - "/bin/sh"
                - "-c"
                - |
                  apk add --no-cache kubectl
                  kubectl scale headroom/example --replicas=0
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: headroom-scaler
  namespace: default # Or your target namespace

---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding # Or RoleBinding if headroom controller is namespace scoped
metadata:
  name: headroom-scaler-rb
  namespace: default
subjects:
  - kind: ServiceAccount
    name: headroom-scaler
    namespace: default
roleRef:
  kind: ClusterRole # Or Role if headroom controller is namespace scoped
  name: headroom-scaler-role
  apiGroup: rbac.authorization.k8s.io

---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole # Or Role if headroom controller is namespace scoped
metadata:
  name: headroom-scaler-role
  namespace: default
rules:
  - apiGroups: ["openfaas.com"]
    resources: ["headrooms", "headrooms/scale"]
    verbs: ["get", "patch", "update"]
```

## Configuration

Specify each parameter using the `--set key=value[,key=value]` argument to `helm install`.

See [values.yaml](./values.yaml) for detailed configuration.

| Parameter | Description | Default |
|-----------|-------------|---------|
| `image` | Container image | See [values.yaml](./values.yaml) |
| `replicas` | Number of controller replicas. It is not recommended to alter this value. | `1` |
| `imagePullPolicy` | Image pull policy | `IfNotPresent` |
| `imagePullSecrets` | Image pull secrets | `[]` |
| `nameOverride` | Override chart name | `""` |
| `fullnameOverride` | Override full chart name | `""` |
| `rbac.create` | Create RBAC resources | `true` |
| `rbac.role` | Role type (Role or ClusterRole) | `ClusterRole` |
| `serviceAccount.create` | Create service account | `true` |
| `serviceAccount.annotations` | Service account annotations | `{}` |
| `serviceAccount.name` | Service account name | `""` |
| `podAnnotations` | Pod annotations | `{}` |
| `podSecurityContext` | Pod security context | `{}` |
| `securityContext` | Container security context | `{}` |
| `resources.limits.cpu` | CPU limit | `250m` |
| `resources.limits.memory` | Memory limit | `256Mi` |
| `resources.requests.cpu` | CPU request | `50m` |
| `resources.requests.memory` | Memory request | `64Mi` |
| `nodeSelector` | Node selector for pod assignment | `{}` |
| `tolerations` | Tolerations for pod assignment | `[]` |
| `affinity` | Affinity rules for pod assignment | `{}` |
| `metrics.enabled` | Enable metrics collection | `false` |
| `leaderElection.enabled` | Enable leader election | `true` |
| `leaderElection.namespace` | Namespace for leader election | `"default"` |
| `logging.level` | Log level | `info` |

## Uninstalling

```bash
helm uninstall headroom-controller
```

Note: This will not remove the CRDs. To remove them:

```bash
kubectl delete crd headrooms.openfaas.com
```
