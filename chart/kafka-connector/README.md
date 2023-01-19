# OpenFaaS Pro Kafka Connector

The [Kafka connector](https://docs.openfaas.com/openfaas-pro/introduction) brings Kafka to OpenFaaS by invoking functions based on Kafka topic annotations.

* See also: [Staying on topic: trigger your OpenFaaS functions with Apache Kafka](https://www.openfaas.com/blog/kafka-connector/)

## Prerequisites

- Purchase a license

  You will need an OpenFaaS License

  Contact us to find out more [openfaas.com/pricing](https://www.openfaas.com/pricing)

- Install OpenFaaS

  You must have a working OpenFaaS installed.

- Self-hosted Kafka with Helm

  For development and testing, you can install [Apache Kafka](https://kafka.apache.org/) using Confluent's [chart](https://github.com/confluentinc/cp-helm-charts). Find out additional options here: [available here](https://github.com/helm/charts/tree/master/incubator/kafka#installing-the-chart)

- Hosted Kafka

  [Aiven](https://aiven.io/) and [Confluent Cloud](https://confluent.cloud/) have both been tested with the Kafka Connector.

Other versions of Kafka like Redpanda should work out of the box.

## Complete walk-through guide

  You can continue with this guide, or start the [walk-through](quickstart.md) for testing purposes.

## Install the Chart

- Create the required secret with your OpenFaaS Pro license code:

```bash
$ kubectl create secret generic \
    -n openfaas \
    openfaas-license \
    --from-file license=$HOME/.openfaas/LICENSE
```

- Add the OpenFaaS chart repo and deploy the `kafka-connector` Pro chart. We recommend installing it in the same namespace as the rest of OpenFaaS

```sh
$ helm repo add openfaas https://openfaas.github.io/faas-netes/
$ helm upgrade kafka-connector openfaas/kafka-connector \
    --install \
    --namespace openfaas
```

> The above command will also update your helm repo to pull in any new releases.

## Install a development version

```sh
$ helm upgrade kafka-connector ./chart/kafka-connector \
    --install \
    --namespace openfaas
```

## Encryption options

1) TLS off (default)
2) TLS on

## Authentication options

1) TLS with SASL using CA from the default trust store
3) TLS with SASL using a custom CA
4) TLS with client certificates

## Configuration

Additional kafka-connector options in `values.yaml`.

| Parameter              | Description                                                                                                        | Default                        |
| ---------------------- | ------------------------------------------------------------------------------------------------------------------ | ------------------------------ |
| `topics`               | Topics to which the connector will bind, provide as a comma-separated list or a single item.                       | `faas-request`                 |
| `brokerHost`           | location of the Kafka brokers.                                                                                     | `kafka`                        |
| `asyncInvocation`      | For long running or slow functions, offload to asychronous function invocations and carry on processing the stream | `false`                        |
| `upstreamTimeout`      | Maximum timeout for upstream function call, must be a Go formatted duration string.                                | `2m`                           |
| `rebuildInterval`      | Interval for rebuilding function to topic map, must be a Go formatted duration string.                             | `30s`                          |
| `gatewayURL`           | The URL for the API gateway.                                                                                       | `http://gateway.openfaas:8080` |
| `printResponse`        | Output the response of calling a function in the logs.                                                             | `true`                         |
| `printResponseBody`    | Output to the logs the response body when calling a function.                                                      | `false`                        |
| `printRequestBody`     | Output to the logs the request body when calling a function.                                                       | `false`                        |
| `fullnameOverride`     | Override the name value used for the Connector Deployment object.                                                  | ``                             |
| `sasl`                 | Enable auth with a SASL username/password                                                                          | `false`                        |
| `brokerPasswordSecret` | Name of secret for SASL password                                                                                   | `kafka-broker-password`        |
| `brokerUsernameSecret` | Name of secret for SASL username                                                                                   | `kafka-broker-username`        |
| `caSecret`             | Name secret for TLS CA - leave empty to disable                                                                    | `kafka-broker-ca`              |
| `certSecret`           | Name secret for TLS client certificate cert - leave empty to disable                                               | `kafka-broker-cert`            |
| `keySecret`            | Name secret for TLS client certificate private key - leave empty to disable                                        | `kafka-broker-key`             |
| `contentType`          | Set a HTTP Content Type during function invocation.                                                                | `""`                           |
| `group`                | Set the Kafka consumer group name.                                                                                 | `""`                           |
| `maxBytes`             | Set the maximum size of messages from the Kafka broker.                                                            | `1024*1024`                    |
| `sessionLogging`       | Enable detailed logging from the consumer group.                                                                   | `"false"`                      |
| `initialOffset`        | Either newest or oldest.                                                                                           | `"oldest"`                     |
| `logs.debug`           | Print debug logs                                                                                                   | `false`                        |
| `logs.format`          | The log encoding format. Supported values: `json` or `console`                                                     | `console`                      |

Specify each parameter using the `--set key=value[,key=value]` argument to `helm install`. See `values.yaml` for the default configuration.

## Removing the kafka-connector

All control plane components can be cleaned up with helm:

```sh
$ helm uninstall -n openfaas kafka-connector
```
