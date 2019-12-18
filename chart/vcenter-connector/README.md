# OpenFaaS vCenter Connector

The [OpenFaaS vCenter connector](https://github.com/openfaas-incubator/openfaas-vcenter-connector) brings vCenter events to OpenFaaS by invoking functions based on vCenter event(s) topic annotations.

## Prerequisites

- Install OpenFaaS

  You must have a working OpenFaaS installation. You can find [instructions in the docs](https://docs.openfaas.com/deployment/kubernetes/#pick-helm-or-yaml-files-for-deployment-a-or-b), including instructions to also install OpenFaaS via Helm.

- Have a working VMware vCenter environment (dev/testing) that you can connect to

## Install the Chart

- Create a Kubernetes secret holding the vCenter credentials to be used by the vcenter-connector

```
kubectl create secret generic vcenter-secrets \
  -n openfaas \
  --from-literal vcenter-username=user \
  --from-literal vcenter-password=pass
```

> **Note:** Replace "user" and "pass" with the actual values. It is a good practice to use a dedicated vCenter service account instead of an administrator role for the vcenter-connector.

- Add the OpenFaaS chart repo and deploy the `vcenter-connector` chart. We recommend installing it in the same namespace as the rest of OpenFaaS

```sh
$ helm repo add openfaas https://openfaas.github.io/faas-netes/
$ helm upgrade vcenter-connector openfaas/vcenter-connector \
    --install \
    --namespace openfaas
```

> **Note:** The above command will also update your helm repo to pull in any new releases.

After you have [verified](#verify-the-installationtroubleshooting) that the installation succeeded you might want to run through an actual [function example](https://github.com/openfaas-incubator/openfaas-vcenter-connector#examples--community).

### Verify the installation/Troubleshooting

1) Check that the `vcenter-connector` deployment in the given Kubernetes namespace (default: "openfaas") is ready, i.e. shows `1/1` pods in running state
2) If you see error messages or `0/1` (note: it might take some time to pull the container image) verify that  
   1) you have followed the [installation](#install-the-chart) steps as outlined above (vCenter secret deployed?)
   2) the status of the corresponding `ReplicaSet` with `kubectl -n openfaas get rs`
   3) the logs of the container don't show any errors with `kubectl -n openfaas logs deploy/vcenter-connector`

## Customizing the Helm Chart

Additional vcenter-connector options in `values.yaml`:

| Parameter            | Description                                                                       | Default                        |
|----------------------|-----------------------------------------------------------------------------------|--------------------------------|
| `resources.requests` | Compute resources Kubernetes will guarantee to this pod                           | `100m (CPU), 120Mi (Memory)`   |
| `vc_k8s_secret`      | vCenter secrets name as stored in Kubernetes                                      | `vcenter-secrets`              |
| `basic_auth`         | Use basic authentication against OpenFaaS gateway                                 | `true`                         |
| `extraArgs.gateway`  | The URL of the OpenFaaS API gateway                                               | `http://gateway.openfaas:8080` |
| `extraArgs.vcenter`  | The URL of the vCenter server                                                     | `http://vcsim.openfaas:8989`   |
| `extraArgs.insecure` | Don't verify vCenter server certificate (remove this parameter to force checking) | `insecure`                     |

Specify each parameter you want to change using the `--set key=value[,key=value]` argument to `helm install`, e.g.:

```
helm install vcenter-connector ./vcenter-connector --namespace openfaas --set extraArgs.vcenter=https://10.160.94.63/sdk
```

See `values.yaml` for detailed configuration.

## Removing the vCenter Connector

The vCenter connector can be cleaned up with the following helm command:

```sh
helm uninstall vcenter-connector --namespace openfaas
```
