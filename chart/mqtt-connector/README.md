# OpenFaaS MQTT Connector chart

See also [mqtt-connector README](https://github.com/openfaas/mqtt-connector)

## Installation

```bash
git clone https://github.com/faas-netes/

cd faas-netes/chart

# Template and apply:

helm template -n openfaas --namespace openfaas mqtt-connector/ | kubectl apply -f -
```

You can watch the Connector logs to see it invoke your functions:

```bash
kubectl logs deployment.apps/openfaas-mqtt-connector -n openfaas -f
```

## Configuration

Configure via `values.yaml`.

| Parameter                | Description                                                               | Example                                     |
| ------------------------ | ------------------------------------------------------------------------- | ------------------------------------------- |
| `topic`                  | A single topic for subscription, install broker once for each topic       | `sample-topic/`                            |
| `broker`                 | A TCP address or websocket for the MQTT broker subscription               | `tcp://emitter:8080`             |
| `clientID`               | An ID to represent this client                                            | `testgoid`                                  |
| `upstream_timeout`       | Maximum timeout for function as (Golang duration)                         | `15s`                                       |
| `rebuild_interval`       | Interval between rebuilding map of functions vs. topics (Golang duration) | `10s`                                       |
| `authPassword`       | Password for broker authentication| `""`                                       |
| `asyncInvoke`       | Invoke using function's asynchronous route | `false`                                       |
| `basic_auth`       | Basic authentication for gateway enabled | `true`                                       |
| `resources`              | Resources requests and limits configuration                               | `requests.memory: "64Mi"`                  |
