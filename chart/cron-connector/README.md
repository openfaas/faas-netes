# OpenFaaS Cron Connector

Schedule function invocations using a cron expression.

## Pre-requisite

You must have a working OpenFaaS installation.

## Install via arkade

```bash
arkade install cron-connector
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

| Parameter          | Description                                      | Default                                  |
|--------------------|--------------------------------------------------|------------------------------------------|
| `image`            | The cron-connector image that should be deployed | `ghcr.io/openfaas/cron-connector:0.3.2`  |
| `gatewayURL`       | The URL for the API gateway.                     | `"http://gateway.openfaas:8080"`         |
| `basicAuth`        | Enable or disable basic auth                     | `true`                                   |
| `asyncInvocation`  | Invoke via the asynchronous function endpoint    | `false`                                  |
| `contentType`      | Set a contentType for all invocations            | `text/plain`                             |

See also: values.yaml

## Remove the cron-connector

All components can be cleaned up with helm:

```bash
helm delete --purge cron-connector
```
