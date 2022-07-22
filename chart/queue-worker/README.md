# JetStream for OpenFaaS queue-worker

Deploy OpenFaaS queues.

## Install an additional queue
You can run the following command from within the `faas-netes` folder, not the chart's folder.

```bash
helm upgrade slow-queue --install chart/queue-worker \
  --namespace openfaas \
  --set maxInflight=5 \
  --set nats.stream.name=slow-queue \
  --set nats.consumer.durableName=slow-queue-workers
```