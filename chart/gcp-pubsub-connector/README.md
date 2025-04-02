# OpenFaaS Pro Google Cloud Pub/Sub Connector

Trigger OpenFaaS functions from Google Cloud Pub/Sub messages.

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

- Create a secret for the Google Cloud Application Default Credentials (ADC).

```bash
$ kubectl create secret generic -n openfaas \
  gcp-pubsub-credentials --from-file gcp-pubsub-credentials=$HOME/gcp-pubsub-credentials.json
```

Ensure you application has the correct [IAM permission to consume messages from Pub/Sub subscriptions](https://cloud.google.com/pubsub/docs/access-control).

## Configure values.yaml

```yaml
# Google cloud project ID
projectID: "openfaas"

# List if Pub/Sub subscriptions the connector should subscribe to.
subscriptions:
  - sub1
  - sub2
```

Use the `subscriptions` parameter to configure a list of Pub/Sub subscriptions to which the connector should subscribe. When the subscriber receives a message, the connector will attempt to invoke any function that has the subscription name listed in its `topic` annotation.

## Install the chart

- Add the OpenFaaS chart repo and deploy the `gcp-pubsub-connector` chart. We recommend installing it in the same namespace as the rest of OpenFaaS

```sh
$ helm repo add openfaas https://openfaas.github.io/faas-netes/
$ helm repo update && \
  helm upgrade gcp-pubsub-connector openfaas/gcp-pubsub-connector \
    --install \
    --namespace openfaas \
    -f values.yaml
```

> The above command will also update your helm repo to pull in any new releases.

## Configuration

Additional gcp-pubsub-connector options in `values.yaml`.

| Parameter              | Description                                                                                                        | Default                        |
| ---------------------- | ------------------------------------------------------------------------------------------------------------------ | ------------------------------ |
| `replicas`             | The number of replicas of this connector. Pub/Sub messages will be load balanced between the connector replicas.   | `1`                            |
| `projectID`            | Google Cloud project Id                                                                                            | `""`                           |
| `subscriptions`        | List if Pub/Sub subscriptions the connector should subscribe to.                                                   | `[]`                           |
| `gcpCredentialsSecret` | Kubernetes secret for the Google Cloud Application Default Credentials (ADC)                                       | `gcp-pubsub-credentials`       |
| `asyncInvocation`      | For long running or slow functions, offload to asychronous function invocations and carry on processing the stream | `false`                        |
| `upstreamTimeout`      | Maximum timeout for upstream function call, must be a Go formatted duration string.                                | `2m`                           |
| `rebuildInterval`      | Interval for rebuilding function to topic map, must be a Go formatted duration string.                             | `30s`                          |
| `gatewayURL`           | The URL for the OpenFaaS gateway.                                                                                  | `http://gateway.openfaas:8080` |
| `printResponse`        | Output the response of calling a function in the logs.                                                             | `true`                         |
| `printResponseBody`    | Output to the logs the response body when calling a function.                                                      | `false`                        |
| `printRequestBody`     | Output to the logs the request body when calling a function.                                                       | `false`                        |
| `fullnameOverride`     | Override the name value used for the Connector Deployment object.                                                  | `""`                           |
| `contentType`          | Set a HTTP Content Type during function invocation.                                                                | `text/plain`                   |
| `logs.debug`           | Print debug logs                                                                                                   | `false`                        |
| `logs.format`          | The log encoding format. Supported values: `json` or `console`                                                     | `console`                      |

Specify each parameter using the `--set key=value[,key=value]` argument to `helm install`. See `values.yaml` for the default configuration.

## Install a development version of the chart

When developing on the chart locally, just specify the path to the chart where you've cloned it:

```sh
$ helm upgrade gcp-pubsub-connector ./chart/gcp-pubsub-connector \
    --install \
    --namespace openfaas \
    -f values.yaml
```

## Removing the gcp-pubsub-connector

All control plane components can be cleaned up with helm:

```sh
$ helm uninstall -n openfaas gcp-pubsub-connector
```
