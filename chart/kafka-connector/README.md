# OpenFaaS Kafka Connector

The [Kafka connector](https://github.com/openfaas-incubator/kafka-connector) brings Kafka to OpenFaaS by invoking functions based on Kafka topic annotations.

## Prerequisites

- Install OpenFaaS

  You must have a working OpenFaaS installation. You can find [instructions in the docs](https://docs.openfaas.com/deployment/kubernetes/#pick-helm-or-yaml-files-for-deployment-a-or-b), including instructions to also install OpenFaaS via Helm.

- Install Kafka

  You must have a working Kafka installation. Confluent has provided various [Helm charts here](https://github.com/confluentinc/cp-helm-charts). Instructions for installing Kafka via Helm are [available here](https://github.com/helm/charts/tree/master/incubator/kafka#installing-the-chart).

## Install the Chart

- Add the OpenFaaS chart repo and deploy the `kafka-connector` chart. We recommend installing it in the same namespace as the rest of OpenFaaS

```sh
$ helm repo add openfaas https://openfaas.github.io/faas-netes/
$ helm upgrade kafka-connector openfaas/kafka-connector \
    --install \
    --namespace openfaas
```

> The above command will also update your helm repo to pull in any new releases.

### Verify the installation

TBD

## Configuration

Additional kafka-connector options in `values.yaml`.

| Parameter          | Description                                                                            | Default                        |
| ------------------ | -------------------------------------------------------------------------------------- | ------------------------------ |
| `upstream_timeout` | Maximum timeout for upstream function call, must be a Go formatted duration string.    | `30s`                          |
| `rebuild_interval` | Interval for rebuilding function to topic map, must be a Go formatted duration string. | `3s`                           |
| `topics`           | Topics to which the connector will bind, provide as a comma-separated list.            | `faas-request`                 |
| `gateway_url`      | The URL for the API gateway.                                                           | `http://gateway.openfaas:8080` |
| `broker_host`      | location of the Kafka brokers.                                                         | `kafka`                        |
| `print_response`   | Output the response of calling a function in the logs.                                 | `true`                         |

Specify each parameter using the `--set key=value[,key=value]` argument to `helm install`.
See values.yaml for detailed configuration.

## Removing the kafka-connector

All control plane components can be cleaned up with helm:

```sh
$ helm delete --purge kafka-connector
```
