# Headroom Controller Helm Chart

A Helm chart for deploying the headroom-controller controller, which manages Kubernetes resource reservations through custom Headroom resources.

## Prerequisites

- Kubernetes 1.19+
- Helm 3.0+

## Installation

### Add the repository

```bash
helm repo add headroom-controller https://openfaas.github.io/faas-netes/
helm repo update

# Install the chart
helm upgrade headroom-controller \
  --install openfaas/headroom-controller \
  --namespace default
```

### Install from local chart

If you are working on a patch for the helm chart, you can deploy it directly from a local folder.

```bash
# Clone the repository
git clone https://github.com/openfaas/faas-netes.git
cd headroom-controller

# Install the chart
helm upgrade install headroom-controller \
  --install./chart/headroom-controller \
  --namespace default
```

## Usage

After installation, you can create Headroom resources:

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

### Check status

```bash
kubectl get headrooms
```

### Scale headroom

```bash
kubectl scale headroom example --replicas=2
```

### Install with custom values

```bash
helm upgrade install headroom-controller \
  --install openfaas/headroom-controller \
  --namespace default \
  --set rbac.role=ClusterRole \
  --set replicaCount=2
```

## Configuration

Specify each parameter using the `--set key=value[,key=value]` argument to `helm install`.

See [values.yaml](./values.yaml) for detailed configuration.

| Parameter | Description | Default |
|-----------|-------------|---------|
| `replicaCount` | Number of controller replicas | `1` |
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
