# OpenFaaS Pro Postgres Connector

The Postgres connector can be used to invoke functions from PostgreSQL database events.

See also: [Trigger functions from Postgres](https://docs.openfaas.com/openfaas-pro/postgres-events/)

## Pre-requisites

- Purchase a license

  You will need an OpenFaaS License

  Contact us to find out more [openfaas.com/pricing](https://www.openfaas.com/pricing)

- Install OpenFaaS

  You must have a working OpenFaaS installed.

- Configure PostgreSQL

  You must have a working PostgreSQL database installed and configured for logical replication, get the specific settings here: [Trigger functions from Postgres](https://docs.openfaas.com/openfaas-pro/postgres-events/)

## Configure your secrets

- Create the required secret with your OpenFaaS Pro license code:

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

Create a `pgconnector.yaml` file with the following custom contents:

```yaml
connectionSecret: "postgresql-connection"

publication: "ofltd"

# filter which tables and events to be notified about
filters: "customer:insert,customer:update"

wal: true
```

## Install the chart

- Add the OpenFaaS chart repo and deploy the `postgres-connector` Pro chart. We recommend installing it in the same namespace as the rest of OpenFaaS

```sh
$ helm repo add openfaas https://openfaas.github.io/faas-netes/
$ helm repo update
$ helm upgrade postgres-connector openfaas/postgres-connector \
    --install \
    --namespace openfaas \
    -f pgconnector.yaml
```

> The above command will also update your helm repo to pull in any new releases.

## Install a development version

```sh
$ helm upgrade postgres-connector ./chart/postgres-connector \
    --install \
    --namespace openfaas \
    -f ./pgconnector.yaml
```

## Configuration

Additional postgres-connector options in `values.yaml`.

| Parameter          | Description                                                                                                        | Default                        |
| ------------------ | ------------------------------------------------------------------------------------------------------------------ | ------------------------------ |
| `wal`              | Start the connector in WAL mode or trigger mode. Set to `false` for trigger mode.                                  | `true`                         |
| `connectionSecret` | The name of the secret containing the connection string                                                            | `postgresql-connection`        |
| `publication`      | The name of the publication to be created in Postgres, we do not recommend changing this value.                    | `ofltd`                        |
| `filters`                | A comma separated list of filters to apply to the stream i.e. `TABLENAME:ACTION`, action can be: `insert`, `delete` or `update`, multiple items can be comma separated such as `T1:insert,T2:delete,T3:insert`  | `""`                           |
| `upstreamTimeout`  | Maximum timeout for upstream function call, must be a Go formatted duration string.                                | `2m`                           |
| `rebuildInterval`  | Interval for rebuilding function to topic map, must be a Go formatted duration string.                             | `30s`                          |
| `asyncInvocation`  | For long running or slow functions, offload to asychronous function invocations and carry on processing the stream | `false`                        |
| `gatewayURL`       | The URL for the API gateway.                                                                                       | `http://gateway.openfaas:8080` |
| `fullnameOverride` | Override the name value used for the Connector Deployment object.                                                  | ``                             |
| `contentType`      | Set a HTTP Content Type during function invocation.                                                                | `""`                           |
| `replicas`         | Number of replicas for the Connector Deployment object, we recommend setting this to `1`                           | `1`                            |
| `resources`        | Resources requests and limits configuration                                                                        | `requests.memory: "64Mi"`      |
| `logs.debug`       | Print debug logs                                                                                                   | `false`                        |
| `logs.format`      | The log encoding format. Supported values: `json` or `console`                                                     | `console`                      |

Specify each parameter using the `--set key=value[,key=value]` argument to `helm install`. See `values.yaml` for the default configuration.

## Removing the postgres-connector

All components can be cleaned up with helm:

```sh
$ helm uninstall -n openfaas postgres-connector
```
