# OpenFaaS Pro RabbitMQ Connector

The RabbitMQ connector can be used to invoke OpenFaaS functions from messages on a RabbitMQ queue.

## Prerequisites

- Purchase a license

  You will need an OpenFaaS License

  Contact us to find out more [openfaas.com/pricing](https://www.openfaas.com/pricing)

- Install OpenFaaS

  You must have a working OpenFaaS installed.

## Configure secrets

- Create the required secret with your OpenFaaS Pro license code:

```bash
$ kubectl create secret generic \
    -n openfaas \
    openfaas-license \
    --from-file license=$HOME/.openfaas/LICENSE
```

## Configure values.yaml

```yaml
rabbitmqURL: "amqp://rabbitmq.rabbitmq.svc.cluster.local:5672"

queues:
  - name: queue1
    durable: true
```

Use the `queues` parameter to configure a list of queues to which the connector should subscribe. When a message is received on a queue, the connector will attempt to invoke any function that has the queue name listed in its `topic` annotation.

Queue configuration:

- `name` - The name of the queue. (Required)
- `durable` - Specifies whether the queue should survive broker restarts.
- `nowait` - Whether to wait for confirmation from the server when declaring the queue.
- `autodelete` - Whether the queue should be deleted automatically when no consumers are connected.
- `exclusive` - Specifies whether the queue should be exclusive to the connection that created it.

By default the connector does not use any form of authentication for the RabbitMQ connection.

Create Kubernetes secrets with username and password to connect to RabbitMQ if required.

```sh
kubectl create secret generic \
  rabbitmq-username \
  -n openfaas \
  --from-file username=./rabbitmq-username.txt

kubectl create secret generic \
  rabbitmq-password \
  -n openfaas \
  --from-file password=./rabbitmq-password.txt
```

To enable username and password authentication add to following parameters with the Kubernetes secret name for the username and password to the `values.yaml`.

```yaml
rabbitmqUsernameSecret: rabbitmq-username
rabbitmqPasswordSecret: rabbitmq-password
```

If your RabbitMQ requires a TLS connection make sure to use `amqps://` as the scheme in the url.

Optionally you can configure a custom CA cert for the connector:

```sh
kubectl create secret generic \
  rabbitmq-ca \
  -n openfaas \
  --from-file ca-cert=./ca-cert.pem
```

`values.yaml`:

```yaml
caSecret: "rabbitmq-ca"
```

## Install the chart

- Add the OpenFaaS chart repo and deploy the `rabbitmq-connector` chart. We recommend installing it in the same namespace as the rest of OpenFaaS

```sh
$ helm repo add openfaas https://openfaas.github.io/faas-netes/
$ helm repo update && \
  helm upgrade rabbitmq-connector openfaas/rabbitmq-connector \
    --install \
    --namespace openfaas \
    -f values.yaml
```

> The above command will also update your helm repo to pull in any new releases.

## Configuration

Additional rabbitmq-connector options in `values.yaml`.

| Parameter                | Description                                                                                                                        | Default                        |
| ------------------------ | ---------------------------------------------------------------------------------------------------------------------------------- | ------------------------------ |
| `topics`                 | A single topic or list of comma separated topics to consume.                                                                       | `faas-request`                 |
| `replicas`               | The number of replicas of this connector, should be set to the size of the partition for the given topic, or a higher lower value. | `1`                            |
| `rabbitmqURL`            | URL used for connecting to a RabbitMQ node.                                                                                        | `amqp://rabbitmq:5672`         |
| `queues`                 | List of queues the connector should subscribe to.                                                                                  | `[]`                           |
| `asyncInvocation`        | For long running or slow functions, offload to asychronous function invocations and carry on processing the stream                 | `false`                        |
| `upstreamTimeout`        | Maximum timeout for upstream function call, must be a Go formatted duration string.                                                | `2m`                           |
| `rebuildInterval`        | Interval for rebuilding function to topic map, must be a Go formatted duration string.                                             | `30s`                          |
| `gatewayURL`             | The URL for the OpenFaaS gateway.                                                                                                  | `http://gateway.openfaas:8080` |
| `printResponse`          | Output the response of calling a function in the logs.                                                                             | `true`                         |
| `printResponseBody`      | Output to the logs the response body when calling a function.                                                                      | `false`                        |
| `printRequestBody`       | Output to the logs the request body when calling a function.                                                                       | `false`                        |
| `fullnameOverride`       | Override the name value used for the Connector Deployment object.                                                                  | `""`                           |
| `rabbitmqPasswordSecret` | Name of the RabbitMQ password secret                                                                                               | `""`                           |
| `rabbitmqUsernameSecret` | Name of the RabbitMQ username secret                                                                                               | `""`                           |
| `caSecret`               | Name secret for TLS CA - leave empty to disable                                                                                    | `""`                           |
| `certSecret`             | Name secret for TLS client certificate cert - leave empty to disable                                                               | `""`                           |
| `keySecret`              | Name secret for TLS client certificate private key - leave empty to disable                                                        | `""`                           |
| `contentType`            | Set a HTTP Content Type during function invocation.                                                                                | `text/plain`                   |
| `logs.debug`             | Print debug logs                                                                                                                   | `false`                        |
| `logs.format`            | The log encoding format. Supported values: `json` or `console`                                                                     | `console`                      |

Specify each parameter using the `--set key=value[,key=value]` argument to `helm install`. See `values.yaml` for the default configuration.

## Install a development version of the chart

When developing on the chart locally, just specify the path to the chart where you've cloned it:

```sh
$ helm upgrade rabbitmq-connector ./chart/rabbitmq-connector \
    --install \
    --namespace openfaas \
    -f values.yaml
```

## Removing the rabbitmq-connector

All control plane components can be cleaned up with helm:

```sh
$ helm uninstall -n openfaas rabbitmq-connector
```
