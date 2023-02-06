# JetStream for OpenFaaS queue-worker

Deploy additional, named OpenFaaS queues.

Named queues requires OpenFaaS Enterprise. Please reach out to us if you're an existing OpenFaaS Pro customer and need this feature.

## Install an additional queue

```bash
helm upgrade slow-queue openfaas/queue-worker \
  --install \
  --namespace openfaas \
  --set maxInflight=5 \
  --set nats.stream.name=slow-queue \
  --set nats.consumer.durableName=slow-queue-workers \
  --set upstreamTimeout=15m
```

When installing from the faas-netes repository on your local computer, whilst making changes or customising the chart, you can run this command from within the `faas-netes` folder.

```bash
helm upgrade slow-queue chart/queue-worker \
  --install \
  --namespace openfaas \
  --set maxInflight=5 \
  --set nats.stream.name=slow-queue \
  --set nats.consumer.durableName=slow-queue-workers \
  --set upstreamTimeout=15m
```

## Configuration
| Parameter | Description | Default |
|-----------|-------------|---------|
| `image` | The jetstream-queue-worker image that should be deployed | See values.yaml |
| `replicas` | Number of queue-worker replicas to create | `1` |
| `maxInflight` | Control the concurrent invocations | `1` |
| `maxWaiting` | Configure the max waiting pulls for the queue-worker JetStream consumer. The value should be at least max_inflight * queue_worker.replicas. Note that this value can not be updated once the consumer is created. | `512` |
| `upstreamTimeout` | Maximum duration of upstream function call | `1m` |
| `maxRetryAttempts` | The amount of times to try sending a message to a function before discarding it |`10` |
| `maxRetryWait` | The maximum amount of time to wait between retries | `120s` |
| `initialRetryWait` | The amount of time to wait for the first retry | `10s` |
| `httpRetryCodes` | A comma-separated list of HTTP status codes which the queue worker will retry when received from a function | `408,429,500,502,503,504` |
| `gateway.host` | The host at which the OpenFaaS gateway can be reached | `http://gateway.openfaas` |
| `insecureTLS` | Enable insecure tls for callbacks | `false` |
| `gateway.port` | The port at which the OpenFaaS gateway can be reached | `8080` |
| `nats.host` | The host at which the NATS JetStream serber can be reached | `nats.openfaas` |
| `nats.port` | The part at which the NATS JetStream server can be reached | `4222` |
| `nats.reconnect.attempt` | Max NATS reconnection attempts | `120` |
| `nats.reconnect.delay` | Time to wait between NATS reconnection attempts | `2s` |
| `nats.stream.name` | Name of the NATS JetStream stream to use | `faas-request` |
| `nats.stream.replicas` | Number of JetStream stream replicas to create | `1` |
| `nats.consumer.durableName` | The name of the NATS JetStream consumer to use | `faas-workers` |
| `logs.debug` | Print debug logs | `false` |
| `logs.format` | The log encoding format. Supported values: `json` or `console` | `console` |