# Headroom Controller Helm Chart

Speed up scaling up and rolling out of workloads by pre-provisioning headroom in your Kubernetes cluster.

## Prerequisites

You do not need to use OpenFaaS to install this chart and use the headroom-controller. It is a standalone component.

- Kubernetes
- Helm
- Some form of autoscaling node group, or Cluster Autoscaler, or Karpenter, etc.

## Demo

Watch a [live demo with the Cluster Autoscaler and K3s](https://www.youtube.com/embed/MHXvhKb6PpA?si=QFEf632Ha3VUESbs):

<iframe width="560" height="315" src="https://www.youtube.com/embed/MHXvhKb6PpA?si=QFEf632Ha3VUESbs" title="YouTube video player" frameborder="0" allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture; web-share" referrerpolicy="strict-origin-when-cross-origin" allowfullscreen></iframe>

## Installation

### Add the repository

```bash
helm repo add openfaas https://openfaas.github.io/faas-netes/ && \
  helm repo update
```

### Install from the chart repository

Install with defaults:

```bash
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
helm upgrade --install headroom-controller \
  openfaas/headroom-controller \
  --namespace default \
  --values values-custom.yaml
```


The chart can also be installed into a custom namespace i.e. `openfaas` or `kube-system` by altering the `--namespace` flag.

### Install from local chart

If you are working on a patch for the helm chart, you can deploy it directly from a local folder.

```bash
# Clone the repository
git clone https://github.com/openfaas/faas-netes.git --depth=1
cd headroom-controller

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

## Configuration

Specify each parameter using the `--set key=value[,key=value]` argument to `helm install`.

See [values.yaml](./values.yaml) for detailed configuration.

| Parameter | Description | Default |
|-----------|-------------|---------|
| `replicaCount` | Number of controller replicas. It is not recommended to alter this value. | `1` |
| `image` | Container image | See [values.yaml](./values.yaml) |
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
