# OpenFaaS Federated Gateway

This proxy is designed to be deployed on a customer's cluster.

The proxy will send on authenticated requests to the local OpenFaaS Standard or CE gateway.

## Licensing and usage

Only for licensed usage by OpenFaaS Enterprise customers with a separate agreement with OpenFaaS Ltd

## Prerequisites

- Install OpenFaaS Standard or CE on a Kubernetes cluster

  You must have at least one target cluster, where you run the OpenFaaS Standard or CE gateway.

## Configure ingress

The proxy needs to be accessible from the control plane, either the Internet with Ingress or Istio, or via an inlets tunnel.

## Configure values.yaml

```yaml
# The audience is the client_id in your OIDC provider which represents this customer
audience: fed-gw.example.com

# The issuer is the root URL for the OIDC provider you're using. It must be accessible on the Internet
# from the federated cluster.
issuer: https://keycloak.example.com/realms/openfaas

# When set to false, only the API and its /system endpoints are allowed to be accessed
# Set allowInvoke to true to allow invoking functions via the /function or /async-function endpoint
allowInvoke: false
```

## Install the chart

- Add the openfaas chart repo and deploy the `federated-gateway` chart. We recommend installing it in the same namespace as the rest of OpenFaaS

```sh
$ helm repo add openfaas https://openfaas.github.io/faas-netes/
$ helm repo update
$ helm upgrade federated-gateway openfaas/federated-gateway \
    --install \
    --namespace openfaas \
    -f ./values.yaml
```

> The above command will also update your helm repo to pull in any new releases.

## Install a development version

```sh
$ helm upgrade federated-gateway ./chart/federated-gateway \
    --install \
    --namespace openfaas \
    -f ./values.yaml
```

## Configuration

Additional federated-gateway options in `values.yaml`.

| Parameter             | Description                                                                                 | Default                        |
| --------------------- | ------------------------------------------------------------------------------------------- | ------------------------------ |
| `issuer`              | The root URL of your IdP                                                                    | `""`                           |
| `audience`            | The client ID for this federated gateway in your IDP                                        | `""`                           |
| `allowInvoke`         | Allow invoking functions via the /function or /async-function endpoint                      | `false`                        |
| `nodeSelector`        | Node labels for pod assignment.                                                             | `{}`                           |
| `affinity`            | Node affinity for pod assignments.                                                          | `{}`                           |
| `tolerations`         | Node tolerations for pod assignment.                                                        | `[]`                           |
| `logs.debug`          | Print debug logs                                                                            | `false`                        |
| `logs.format`         | The log encoding format. Supported values: `json` or `console`                              | `console`                      |

Specify each parameter using the `--set key=value[,key=value]` argument to `helm install`. See `values.yaml` for the default configuration.

## Removing the federated-gateway

All components can be cleaned up with helm:

```sh
$ helm uninstall -n openfaas federated-gateway
```
