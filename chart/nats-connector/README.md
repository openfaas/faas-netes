# OpenFaaS Nats Connector

The [Nats connector](https://github.com/openfaas-incubator/nats-connector) invokes functions based on Nats topic annotations.

## Prerequisites

- Install OpenFaaS

  You must have a working OpenFaaS installation. You can find [instructions in the docs](https://docs.openfaas.com/deployment/kubernetes/#pick-helm-or-yaml-files-for-deployment-a-or-b), including instructions to also install OpenFaaS via Helm.


## Install the Chart

- Add the OpenFaaS chart repo and deploy the `nats-connector` chart. We recommend installing it in the same namespace as the rest of OpenFaaS

```sh
$ helm repo add openfaas https://openfaas.github.io/faas-netes/
$ helm upgrade nats-connector openfaas/nats-connector \
    --install \
    --namespace openfaas
```

> The above command will also update your helm repo to pull in any new releases.

## Configuration

Additional nats-connector options in `values.yaml`.

| Parameter                | Description                                                                            | Default                        |
| ------------------------ | -------------------------------------------------------------------------------------- | ------------------------------ |
| `gateway_url`       | OpenFaaS gateway URL   | `http://gateway.openfaas:8080`                          |
| `basic_auth`       | Enable basic auth for the gateway   | `true`                          |
| `upstream_timeout`       | Maximum timeout for upstream function call, must be a Go formatted duration string.    | `30s`                          |
| `rebuild_interval`       | Interval for rebuilding function to topic map, must be a Go formatted duration string. | `5s`                           |
| `topics`                 | Topics to which the connector will bind, provide as a comma-separated list.            | `nats-test,`                 |
| `topic_delimiter`                 | Delimiter for topics list.            | `,`                 |
| `gateway_url`            | The URL for the API gateway.                                                           | `http://gateway:8080"` |
| `broker_host`            | location of the nats brokers.                                                         | `nats`                        |
| `print_response`         | Output the response of calling a function in the logs.                                 | `true`                         |
| `print_response_body`         | Output to the logs the response body when calling a function.                                 | `false`                         |
| `fullnameOverride`       | Override the name value used for the Connector Deployment object.                      | ``                             |

Specify each parameter using the `--set key=value[,key=value]` argument to `helm install`.
See values.yaml for detailed configuration.

## Removing the nats-connector

All control plane components can be cleaned up with helm:

For Helm 2

```sh
$ helm delete --purge nats-connector
```

For Helm 3

```sh
$ helm uninstall nats-connector
```