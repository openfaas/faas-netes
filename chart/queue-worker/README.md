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
| `image` | The queue-worker image that should be deployed | See values.yaml |
| `replicas` | Number of queue-worker replicas to create | `1` |
| `queueName` | Name of the queue | `faas-request` |
| `mode` | Queue operation mode: `static` or `function` | `static` |
| `maxInflight` | Control the concurrent invocations | `1` |
| `queuePartitions` | Number of queue partitions | `1` |
| `partition` | Queue partition number this queue should subscribe to | `0` |
| `consumer.inactiveThreshold` | If a function is inactive (has no invocations) for longer than this threshold its consumer will be removed to save resources | `30s` |
| `consumer.pullMaxMessages` | PullMaxMessages limits the number of messages to be buffered per consumer. Leave empty to use optimized default for the selected queue mode | `` |
| `upstreamTimeout` | Maximum duration of upstream function call | `1m` |
| `maxRetryAttempts` | The amount of times to try sending a message to a function before discarding it | `10` |
| `maxRetryWait` | The maximum amount of time to wait between retries | `120s` |
| `initialRetryWait` | The amount of time to wait for the first retry | `10s` |
| `httpRetryCodes` | A comma-separated list of HTTP status codes which the queue worker will retry when received from a function | `408,429,500,502,503,504` |
| `backoff` | The backoff algorithm used for retries. Must be one off `exponential`, `full` or `equal`| `exponential` |
| `gateway.host` | The host at which the OpenFaaS gateway can be reached | `http://gateway.openfaas` |
| `gateway.port` | The port at which the OpenFaaS gateway can be reached | `8080` |
| `insecureTLS` | Enable insecure tls for callbacks | `false` |
| `nats.host` | The host at which the NATS JetStream server can be reached | `nats.openfaas` |
| `nats.port` | The port at which the NATS JetStream server can be reached | `4222` |
| `nats.stream.name` | Name of the NATS JetStream stream to use | `faas-request` |
| `nats.stream.replicas` | Number of JetStream stream replicas to create | `1` |
| `nats.consumer.durableName` | The name of the NATS JetStream consumer to use | `faas-workers` |
| `nats.consumer.ackWait` | AckWait configures how long the NATS waits for an acknowledgement before redelivering the message| `30s` |
| `logs.debug` | Print debug logs | `false` |
| `logs.format` | The log encoding format. Supported values: `json` or `console` | `console` |
| `resources.requests.memory` | Memory resource request | `120Mi` |
| `resources.requests.cpu` | CPU resource request | `50m` |
| `imagePullPolicy` | Image pull policy | `IfNotPresent` |
| `nodeSelector` | [NodeSelector](https://kubernetes.io/docs/concepts/scheduling-eviction/assign-pod-node/) for deployment | `{}` |
| `tolerations` | [Tolerations](https://kubernetes.io/docs/concepts/scheduling-eviction/taint-and-toleration/) for deployment | `[]` |
| `affinity` | [Aaffinity](https://kubernetes.io/docs/concepts/scheduling-eviction/assign-pod-node/) rules for deployment | `{}` |
