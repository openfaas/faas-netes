# OpenFaaS Cron Connector

Schedule function invocations using a cron expression.

OpenFaaS Pro enables compatibility with other connectors at the same time. For the CE version to work, the topic must only contain one value `cron-function`.

## Pre-requisite

You must have a working OpenFaaS installation.

## Install via arkade

```bash
arkade install cron-connector

# Install for Pro

arkade install cron-connector --set openfaasPro=true
```

## Configure for token-based auth with [OpenFaaS IAM](https://docs.openfaas.com/openfaas-pro/iam/overview/)

```yaml
openfaasPro: true

iam:
  enabled: true
  systemIssuer:
  # URL for the OpenFaaS OpenID connect provider for system components. This is usually the public url of the gateway.
    url: "https://gateway.example.com"
  # Scope of access for the connector.
  # By default it operates on all function namespaces.
  resource: ["*"]
```

## Install via the published chart

- Add OpenFaaS chart repo to helm and install using the following command.

```bash
# Add OpenFaaS chart repo
$ helm repo add openfaas https://openfaas.github.io/faas-netes/

$ helm upgrade --install \
cron-connector openfaas/cron-connector \
    --namespace openfaas
```

## Deploying using local version of the chart

```bash
git clone https://github.com/openfaas/faas-netes.git

cd faas-netes

helm upgrade --install --namespace openfaas \
  cron-connector \
  ./chart/cron-connector
```

## Configuration options

| Parameter           | Description                                                                  | Default                          |
| ------------------- | ---------------------------------------------------------------------------- | -------------------------------- |
| `openfaasPro`       | Enable or disable OpenFaaS Pro features                                      | `false`                          |
| `openfaasPro.image` | The OpenFaaS Pro image that should be deployed                               | See values.yaml                  |
| `image`             | The cron-connector image that should be deployed                             | See values.yaml                  |
| `gatewayURL`        | The URL for the API gateway.                                                 | `"http://gateway.openfaas:8080"` |
| `basicAuth`         | Enable or disable basic auth                                                 | `true`                           |
| `asyncInvocation`   | Invoke via the asynchronous function endpoint                                | `false`                          |
| `iam.enabled` | Enable token-based authentication for use with OpenFaaS IAM | `false` |
| `iam.systemIssuer.url` | URL for the OpenFaaS OpenID connect provider for system components. This is usually the public url of the gateway. | `"https://gateway.example.com"` |
| `iam.kubernetesIssuer.url` | URL for the Kubernetes service account issuer. | `"https://kubernetes.default.svc.cluster.local"` |
| `iam.resource` | Permission scope for the connecter. By default the connecter has access to all namespaces. | `["*"]` |

| `contentType`       | Set a contentType for all invocations                                        | `text/plain`                     |
| `logs.debug`        | Print debug logs (pro feature)                                               | `false`                          |
| `logs.format`       | The log encoding format. Supported values: `json` or `console` (pro feature) | `console`                        |

See also: values.yaml

## Remove the cron-connector

All components can be cleaned up with helm:

```bash
helm uninstall -n openfaas cron-connector
```
