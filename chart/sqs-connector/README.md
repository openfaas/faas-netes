# OpenFaaS Pro SQS Connector

The SQS connector can be used to invoke functions from an AWS SQS queue.

## Prerequisites

- Purchase a license

  You will need an OpenFaaS License

  Contact us to find out more [openfaas.com/pricing](https://www.openfaas.com/pricing)

- Install OpenFaaS

  You must have a working OpenFaaS installed.

## Configure your secrets

- Create the required secret with your OpenFaaS Pro license code:

```bash
$ kubectl create secret generic \
    -n openfaas \
    openfaas-license \
    --from-file license=$HOME/.openfaas/LICENSE
```

- If not using AWS ambient credentials, create a secret:

```bash
$ kubectl create secret generic -n openfaas \
  aws-sqs-credentials --from-file aws-sqs-credentials=$HOME/sqs-credentials.txt
```

## Configure values.yaml

```yaml
replicas: 1

queueURL: "https://"

# Set to empty string if using AWS ambient credentials
awsCredentialsSecret: aws-sqs-credentials

# Set the region
awsDefaultRegion: eu-west-1

# Maximum time to keep message hidden from other processors whilst
# executing function
visibilityTimeout: 30s

# Time to wait between polling SQS queue for messages.
waitTime: 20s

# Maximum messages to fetch at once - between 1-10
maxMessages: 1
```

## Install the chart

- Add the OpenFaaS chart repo and deploy the `sqs-connector` Pro chart. We recommend installing it in the same namespace as the rest of OpenFaaS

```sh
$ helm repo add openfaas https://openfaas.github.io/faas-netes/
$ helm upgrade sqs-connector openfaas/sqs-connector \
    --install \
    --namespace openfaas
```

> The above command will also update your helm repo to pull in any new releases.

## Install a development version

```sh
$ helm upgrade sqs-connector ./chart/sqs-connector \
    --install \
    --namespace openfaas
    -f ./values.yaml
```

## HTTP Headers

Your function will receive these headers:

* `X-SQS-Message-ID` - the ID of the message being processed.
* `X-SQS-Queue-URL` - the URL of the queue.
* `Content-Type` - if configured for the controller.

## Configuration

Additional sqs-connector options in `values.yaml`.

| Parameter                | Description                                                                            | Default                        |
| ------------------------ | -------------------------------------------------------------------------------------- | ------------------------------ |
| `asyncInvocation`        | For long running or slow functions, offload to asychronous function invocations and carry on processing the stream | `false`   |
| `upstreamTimeout`        | Maximum timeout for upstream function call, must be a Go formatted duration string.    | `2m`                          |
| `rebuildInterval`        | Interval for rebuilding function to topic map, must be a Go formatted duration string. | `30s`                           |
| `gatewayURL`             | The URL for the API gateway.                                                           | `http://gateway.openfaas:8080` |
| `printResponse`          | Output the response of calling a function in the logs.                                 | `true`                         |
| `printResponseBody`      | Output to the logs the response body when calling a function.                          | `false`                        |
| `printRequestBody`       | Output to the logs the request body when calling a function.                           | `false`                        |
| `fullnameOverride`       | Override the name value used for the Connector Deployment object.                      | ``                             |
| `contentType`            | Set a HTTP Content Type during function invocation.                                    | `""`                           |
| `resources`              | Resources requests and limits configuration                               | `requests.memory: "64Mi"`                  |
| `logs.debug`           | Print debug logs                                                                                                   | `false`                        |
| `logs.format`          | The log encoding format. Supported values: `json` or `console`                                                     | `console`                      |

Specify each parameter using the `--set key=value[,key=value]` argument to `helm install`. See `values.yaml` for the default configuration.

## Removing the sqs-connector

All control plane components can be cleaned up with helm:

```sh
$ helm uninstall -n openfaas sqs-connector
```
