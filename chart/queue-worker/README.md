# JetStream for OpenFaaS queue-worker

Use this chart to [deploy additional queues and queue-workers](https://docs.openfaas.com/reference/async/#dedicated-queues), for functions which need to be separated out from the default queue.

> Note: This chart requires OpenFaaS for Enterprises. Please [reach out to us](https://openfaas.com/pricing) to discuss upgrading your installation.

Reasons why you may want an additional queue-worker:

* You run workloads for your own product, and for tenants, and want to separate the two
* You want to run some queues with different settings, such as `maxInflight` (perhaps a function has a specific amount of requests it can handle at once), `upstreamTimeout` (longer timeouts), `mode` (static/function based consumers)
* Certain long-running invocations may "clog up" the queue and starve requests from other functions, so you want to run those separately

## Install an additional queue-worker

```bash
helm upgrade --install \
  slow-fns openfaas/queue-worker \
  --namespace openfaas \
  --set maxInflight=5 \
  --set mode=static \
  --set nats.stream.name=slow-fns \
  --set nats.consumer.durableName=slow-fns-workers \
  --set upstreamTimeout=15m
```

It's recommended to take the NAME of the queue i.e. `slow-fns` and then use it as prefix in the following way for the above configuration:

* `nats.stream.name` - NAME`-requests`
* `nats.consumer.durableName` - NAME`-workers`

As an alternative to using `--set`, you could also write your own YAML file. Below is the equivalent configuration in a values.yaml file, for instance `values-slow-fns.yaml`:

```yaml
maxInflight: 5
mode: static
nats:
  stream:
    name: slow-fns
  consumer:
    durableName: slow-fns-workers
upstreamTimeout: 15m
```

Then pass `-f ./values-slow-fns.yaml` to the `helm upgrade --install` command instead of the `--set` flags.

The chart will append a suffix of `-queue-worker` to the release name given above, so the name of the queue-worker will be `slow-fns-queue-worker`.

Upon start-up, the queue-worker will create a NATS JetStream stream named `slow-fns`. Depending on the queue mode either one static consumer will be created, or if the mode is set to `function`, a consumer will be created for each function that has invocations pending.

Remember to set the `com.openfaas.queue` annotation for your functions, so that their requests get submitted to the correct queue (NATS JetStream Stream).

For example:

```bash
# Runs on the slow queue
faas-cli store deploy sleep --name slow-fn --annotation com.openfaas.queue=slow-fns

faas-cli invoke --async slow-fn <<< ""

# Runs on the default queue
faas-cli store deploy env --name fast-fn

faas-cli invoke --async fast-fn <<< ""
```

To remove a queue-worker, run:

```bash
helm uninstall --namespace openfaas \
  slow-fns
```

## Install a development version of the chart

The below instructions are for development on the chart itself, or if you want to test changes to the chart's templates that have not yet been released.

Checkout the code, and cd to the folder:

```bash
git clone https://github.com/openfaas/faas-netes --depth=1
cd faas-netes/chart/queue-worker
```

```bash
helm upgrade --install \
  slow-fns ./ \
  --namespace openfaas \
  --set maxInflight=5 \
  --set mode=static \
  --set nats.stream.name=slow-fns \
  --set nats.consumer.durableName=slow-fns-workers \
  --set upstreamTimeout=15m
```

## Configuration

| Parameter | Description | Default |
|-----------|-------------|---------|
| `image` | The queue-worker image that should be deployed | See values.yaml |
| `replicas` | Number of queue-worker replicas to create | `1` |
| `queueName` | Name of the queue if you want it to be different to the stream name | `""` - when empty, defaults to `nats.stream.name` |
| `mode` | Queue operation mode: `static` (OpenFaaS Standard) or `function` (requires OpenFaaS for Enterprises) | `static` |
| `maxInflight` | Control the concurrent invocations | `1` |
| `adaptiveConcurrency` | Enable adaptive concurrency limiting for functions based on 429 response. This setting only takes effect when `mode` is set to `function`. | `true` |
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
