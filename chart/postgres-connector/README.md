# OpenFaaS Pro Postgres Connector

The Postgres connector can be used to invoke functions from PostgreSQL database events.

## Prerequisites

- Obtain a license or trial

  You will need an OpenFaaS Premium subscription to access PRO features.

  Contact us to find out more and to start a free trial at: [openfaas.com/support](https://www.openfaas.com/support)

- Install OpenFaaS

  You must have a working OpenFaaS installed.

## Configure your secrets

- Create the required secret with your OpenFaaS PRO license code:

```bash
$ kubectl create secret generic \
    -n openfaas \
    openfaas-license \
    --from-file license=$HOME/.openfaas/LICENSE
```

- Create a secret with your PostgreSQL connection string:

```bash
$ kubectl create secret generic -n openfaas \
  postgresql-connection --from-file postgresql-connection=$HOME/postgresql-connection.txt
```

## Configure values.yaml

```yaml
replicas: 1

connectionFileSecret: "postgres-connection-file"

publication: "ofltd"

# filter which tables and events to be notified about
filters: "customers:insert,customers:update"
```

## Install the chart

- Add the OpenFaaS chart repo and deploy the `postgres-connector` PRO chart. We recommend installing it in the same namespace as the rest of OpenFaaS

```sh
$ helm repo add openfaas https://openfaas.github.io/faas-netes/
$ helm upgrade postgres-connector openfaas/postgres-connector \
    --install \
    --namespace openfaas
```

> The above command will also update your helm repo to pull in any new releases.

## Install a development version

```sh
$ helm upgrade postgres-connector ./chart/postgres-connector \
    --install \
    --namespace openfaas
    -f ./values.yaml
```

## Configuration

Additional postgres-connector options in `values.yaml`.

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

Specify each parameter using the `--set key=value[,key=value]` argument to `helm install`. See `values.yaml` for the default configuration.

## Removing the postgres-connector

All components can be cleaned up with helm:

```sh
$ helm uninstall -n openfaas postgres-connector
```
