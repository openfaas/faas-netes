# OpenFaaS Kafka Connector

The [Kafka connector](https://github.com/openfaas-incubator/kafka-connector) brings Kafka to OpenFaaS by invoking functions based on Kafka topic annotations.

* See also: [Staying on topic: trigger your OpenFaaS functions with Apache Kafka](https://www.openfaas.com/blog/kafka-connector/)

## Prerequisites

- Install OpenFaaS

  You must have a working OpenFaaS installation. You can find [instructions in the docs](https://docs.openfaas.com/deployment/kubernetes/#pick-helm-or-yaml-files-for-deployment-a-or-b), including instructions to also install OpenFaaS via Helm.

- Install Kafka (dev/testing)

  You can install [Apache Kafka](https://kafka.apache.org/) with the [wurstmeister Docker images](https://github.com/wurstmeister/kafka-docker) from the [kafka-connector repo](https://github.com/openfaas-incubator/kafka-connector).

  ```sh
  $ git clone https://github.com/openfaas-incubator/kafka-connector
  $ cd kafka-connector/yaml/kubernetes
  $ kubectl apply -f \
   kafka-broker-dep.yml,kafka-broker-svc.yml,zookeeper-dep.yaml,zookeeper-svc.yaml
  ```

- Install Kafka (production)

  You may already have a working Kafka installation, but if not you can provision it using a production-grade helm chart. Confluent has provided various [Helm charts here](https://github.com/confluentinc/cp-helm-charts). Instructions for installing Apache Kafka via Helm are [available here](https://github.com/helm/charts/tree/master/incubator/kafka#installing-the-chart).

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

| Parameter                | Description                                                                            | Default                        |
| ------------------------ | -------------------------------------------------------------------------------------- | ------------------------------ |
| `upstream_timeout`       | Maximum timeout for upstream function call, must be a Go formatted duration string.    | `30s`                          |
| `rebuild_interval`       | Interval for rebuilding function to topic map, must be a Go formatted duration string. | `3s`                           |
| `topics`                 | Topics to which the connector will bind, provide as a comma-separated list.            | `faas-request`                 |
| `gateway_url`            | The URL for the API gateway.                                                           | `http://gateway.openfaas:8080` |
| `broker_host`            | location of the Kafka brokers.                                                         | `kafka`                        |
| `print_response`         | Output the response of calling a function in the logs.                                 | `true`                         |
| `print_response_body`         | Output to the logs the response body when calling a function.                                 | `false`                         |
| `fullnameOverride`       | Override the name value used for the Connector Deployment object.                      | ``                             |

Specify each parameter using the `--set key=value[,key=value]` argument to `helm install`.
See values.yaml for detailed configuration.

## Removing the kafka-connector

All control plane components can be cleaned up with helm:

```sh
$ helm delete --purge kafka-connector
```
