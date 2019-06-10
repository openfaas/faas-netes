# Using helm chart for deployment

```bash
git clone https://github.com/zeerorg/cron-connector.git

cd cron-connector

helm upgrade --namespace openfaas \
  --install cron-connector \
  ./chart/cron-connector
```

## Configuration

cron-connector options in values.yaml

| Parameter     | Description                                      | Default                          |
|---------------|--------------------------------------------------|----------------------------------|
| `image`       | The cron-connector image that should be deployed | `zeerorg/cron-connector:v0.2.2`  |
| `gateway_url` | The URL for the API gateway.                     | `"http://gateway.openfaas:8080"` |
| `basic_auth`  | Enable or disable basic auth                     | `true`                           |

## Removing the cron-connector

All components can be cleaned up with helm:

```bash
helm delete --purge cron-connector
```
