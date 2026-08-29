# OpenFaaS Pro Kafka Connector

The Kafka connector brings Kafka to OpenFaaS by invoking functions based on Kafka topic annotations.

See also: [Docs & tutorials for the Kafka connector](https://docs.openfaas.com/openfaas-pro/introduction)

## Prerequisites

- Obtain an license for OpenFaaS

  You will need an OpenFaaS license, [find out more](https://www.openfaas.com/pricing)

- Install OpenFaaS

  You must have a working OpenFaaS installation.

- Option 1 - hosted Kafka

  [Aiven](https://aiven.io/) and [Confluent Cloud](https://confluent.cloud/) have both been tested with the Kafka Connector.

  Other hosted Kafka offerings such as Kafka like Redpanda should also work.

- Option 2 - Self-hosted Kafka

  For development and testing, you can install [Apache Kafka](https://kafka.apache.org/) using Confluent's [chart](https://github.com/confluentinc/cp-helm-charts). Find out additional options here: [available here](https://github.com/helm/charts/tree/master/incubator/kafka#installing-the-chart)

- Option 3 - Use an existing Kafka deployment

  If you already have Kafka deployed, you'll all the usual details such as whether TLS is enabled along with any required certs, what kind of authentication is being used, what the topic names are, and the partition size for each.

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
```

Prepare a custom [values.yaml](values.yaml) with:

- brokerHosts - comma separted list of host:port
- topics - the topics to subscribe to
- replicas - this should match the partition size, so if the size is 3, set this to 3

Then you will need to read up on the encryption and authentication options and update the settings accordingly.

```sh
$ helm repo update && \
  helm upgrade kafka-connector openfaas/kafka-connector \
    --install \
    --namespace openfaas \
    -f ./custom-values.yaml
```

> The above command will also update your helm repo to pull in any new releases.

## Encryption options

1. TLS off (default)
2. TLS on

## Authentication options

1. TLS with SASL using CA from the default trust store
2. TLS with SASL using a custom CA
3. TLS with client certificates

## Async invocations

The connector can be configured to invoke function asynchronously. This lets you use [OpenFaaS async](https://docs.openfaas.com/reference/async/) features like retries.
To prevent the connector from consuming all Kafka messages at once and submitting them to the OpenFaaS async queue a limit on the number of inflight async invocations can be configured.

Configure the connector for async invocations:

```yaml
# Invoke functions asynchronously.
asyncInvocation: true

async:
  # Limit the number of inflight async invocations for the connector.
  # A value of 0 indicates no concurrency limit.
  maxInflight: 0

# Configure an externally-managed NATS server.
# NATS is used for async invocations and is required when
# setting the 'async.maxInflight' parameter to a value other than 0.
# By default the OpenFaaS embedded nats deployment is used.
# These values should be identical to the configuration in your OpenFaaS deployment values.yaml file
# when external nats is enabled.
nats:
  external:
    enabled: false
    host: ""
    port: ""
```

### Reset the inflight concurrency counter

If the inflight counter gets out if sync for some reason, e.g a misconfiguration, network issues, it can be forcefully reset.
The connecter checks if a Lease object exists on startup and resets the counter if the Lease does not exist.

Remove the lease and restart the connector to reset the counter.

1. Remove the lease

```sh
$ kubectl get lease -n openfaas

NAME              HOLDER   AGE
kafka-connector            18m
```

```sh
kubectl delete lease kafka-connector -n openfaas
```

2. Restart the connector

```sh
kubectl rollout restart deploy/kafka-connector -n openfaas
```

## Configuration

Additional kafka-connector options in `values.yaml`.

| Parameter               | Description                                                                                                                        | Default                        |
| ----------------------- | ---------------------------------------------------------------------------------------------------------------------------------- | ------------------------------ |
| `topics`                | A single topic or list of comma separated topics to consume.                                                                       | `faas-request`                 |
| `replicas`              | The number of replicas of this connector, should be set to the size of the partition for the given topic, or a higher lower value. | `1`                            |
| `brokerHosts`           | Host and port for the Kafka bootstrap server, multiple servers can be specified as a comma-separated list.                         | `kafka:9092`                   |
| `asyncInvocation`       | Invoke function asychronously and carry on processing the stream                                                                   | `false`                        |
| `async.maxInflight`     | Limit the number of inflight async invocations for the connector. A value of 0 indicates no concurrency limit.                     | `0`                            |
| `nats.external.enabled` | Whether to use an externally-managed NATS server.                                                                                  | `false`                        |
| `nats.external.host`    | The host at which the externally-managed NATS server can be reached                                                                | `""`                           |
| `nats.external.port`    | The port at which the externally-managed NATS server can be reached                                                                | `""`                           |
| `upstreamTimeout`       | Maximum timeout for upstream function call, must be a Go formatted duration string.                                                | `2m`                           |
| `rebuildInterval`       | Interval for rebuilding function to topic map, must be a Go formatted duration string.                                             | `30s`                          |
| `gatewayURL`            | The URL for the API gateway.                                                                                                       | `http://gateway.openfaas:8080` |
| `printResponse`         | Output the response of calling a function in the logs.                                                                             | `true`                         |
| `printResponseBody`     | Output to the logs the response body when calling a function.                                                                      | `false`                        |
| `printRequestBody`      | Output to the logs the request body when calling a function.                                                                       | `false`                        |
| `fullnameOverride`      | Override the name value used for the Connector Deployment object.                                                                  | `""`                           |
| `tls`                   | Connect to the broker server(s) using TLS encryption                                                                               | `true`                         |
| `sasl`                  | Enable auth with a SASL username/password                                                                                          | `false`                        |
| `brokerPasswordSecret`  | Name of secret for SASL password                                                                                                   | `kafka-broker-password`        |
| `brokerUsernameSecret`  | Name of secret for SASL username                                                                                                   | `kafka-broker-username`        |
| `caSecret`              | Name secret for TLS CA - leave empty to disable                                                                                    | `kafka-broker-ca`              |
| `certSecret`            | Name secret for TLS client certificate cert - leave empty to disable                                                               | `kafka-broker-cert`            |
| `keySecret`             | Name secret for TLS client certificate private key - leave empty to disable                                                        | `kafka-broker-key`             |
| `contentType`           | Set a HTTP Content Type during function invocation.                                                                                | `""`                           |
| `group`                 | Set the Kafka consumer group name.                                                                                                 | `""`                           |
| `maxBytes`              | Set the maximum size of messages from the Kafka broker.                                                                            | `1024*1024`                    |
| `sessionLogging`        | Enable detailed logging from the consumer group.                                                                                   | `"false"`                      |
| `initialOffset`         | Either newest or oldest.                                                                                                           | `"oldest"`                     |
| `logs.debug`            | Print debug logs                                                                                                                   | `false`                        |
| `logs.format`           | The log encoding format. Supported values: `json` or `console`                                                                     | `console`                      |

Specify each parameter using the `--set key=value[,key=value]` argument to `helm install`. See `values.yaml` for the default configuration.

## Install a development version of the chart

When developing on the chart locally, just specify the path to the chart where you've cloned it:

```sh
$ helm upgrade kafka-connector ./chart/kafka-connector \
    --install \
    --namespace openfaas \
    -f values.yaml
```

## Removing the kafka-connector

All control plane components can be cleaned up with helm:

```sh
$ helm uninstall -n openfaas kafka-connector
```
