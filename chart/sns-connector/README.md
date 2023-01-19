# OpenFaaS Pro SQS Connector

The SNS connector can be used to invoke functions from an AWS SNS messages.

See also: [Trigger functions from AWS SNS messages](https://docs.openfaas.com/openfaas-pro/sns-events/)

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

- Create an AWS credentials secret:

```bash
$ kubectl create secret generic -n openfaas \
  aws-sns-credentials --from-file aws-sns-credentials=$HOME/sns-credentials.txt
```

You can configure permissions using a dedicated IAM user. The user needs a policy that grants access to the `Subscribe` and `ConfirmSubscription` actions. Optionally you can also limit the topics it has access to. For more information see: [Using identity-based policies with Amazon SNS](https://docs.aws.amazon.com/sns/latest/dg/sns-using-identity-based-policies.html)

## Configure ingress

To receive http calls from AWS SNS the callback url has to be publicly accessible.

The below instructions show how to set up Ingress with a TLS certificate using Ingress Nginx. You can also use any other ingress-controller, inlets-pro or an Istio Gateway. Reach out to us if you need a hand.

Install [cert-manager](https://cert-manager.io/docs/), which is used to manage TLS certificates.

You can use Helm, or [arkade](https://github.com/alexellis/arkade):

```bash
arkade install cert-manager
```

Install ingress-nginx using arkade or Helm:

```bash
arkade install ingress-nginx
```

Create an ACME certificate issuer:

```bash
export EMAIL="mail@example.com"

cat > issuer-prod.yaml <<EOF
apiVersion: cert-manager.io/v1
kind: Issuer
metadata:
  name: letsencrypt-prod
  namespace: openfaas
spec:
  acme:
    server: https://acme-v02.api.letsencrypt.org/directory
    email: $EMAIL
    privateKeySecretRef:
      name: letsencrypt-prod
    solvers:
    - http01:
        ingress:
          class: nginx
EOF
```

```bash
kubectl apply -f issuer-prod.yaml
```

Create an ingress record for the sns-connector:

```bash
export DOMAIN="sns.example.com"

cat > ingress.yaml <<EOF
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: sns-connector
  namespace: openfaas
  annotations:
    kubernetes.io/ingress.class: nginx
    cert-manager.io/issuer: letsencrypt-prod
  labels:
    app: sns-connector
spec:
  tls:
  - hosts:
    - $DOMAIN
    secretName: sns-connector-cert
  rules:
  - host: $DOMAIN
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: sns-connector
            port:
              number: 8080
EOF
```

Apply the Ingress resource:

```bash
kubectl apply -f ingress.yaml
```

## Configure values.yaml

```yaml
# Public callback URL for subscriptions
callbackURL: "http://sns.example.com/callback"

# SNS topic ARN
topicARN: "arn:aws:sns:us-east-1:123456789012:of-event"

# AWS shared credentials file:
awsCredentialsSecret: aws-sns-credentials

awsRegion: us-east-1
```

## Install the chart

- Add the OpenFaaS chart repo and deploy the `sns-connector` chart. We recommend installing it in the same namespace as the rest of OpenFaaS

```sh
$ helm repo add openfaas https://openfaas.github.io/faas-netes/
$ helm upgrade sns-connector openfaas/sns-connector \
    --install \
    --namespace openfaas
```

> The above command will also update your helm repo to pull in any new releases.

## Install a development version

```sh
$ helm upgrade sns-connector ./chart/sns-connector \
    --install \
    --namespace openfaas
    -f ./values.yaml
```

## Configuration

Additional sns-connector options in `values.yaml`.

| Parameter              | Description                                                                                                        | Default                        |
| ---------------------- | ------------------------------------------------------------------------------------------------------------------ | ------------------------------ |
| `callbackURL`          | Public callback URL for subscriptions                                                                              | `""`                           |
| `topicARN`             | Amazon SNS topic ARN                                                                                               | `""`                           |
| `awsCredentialsSecret` | Kubernetes secret for the AWS shared credentials file                                                              | `aws-sns-credentials`          |
| `awsRegion`            | The AWS region                                                                                                     | `eu-west-1`                    |
| `asyncInvocation`      | For long running or slow functions, offload to asychronous function invocations and carry on processing the stream | `false`                        |
| `upstreamTimeout`      | Maximum timeout for upstream function call, must be a Go formatted duration string.                                | `2m`                           |
| `rebuildInterval`      | Interval for rebuilding function to topic map, must be a Go formatted duration string.                             | `30s`                          |
| `gatewayURL`           | The URL for the API gateway.                                                                                       | `http://gateway.openfaas:8080` |
| `printResponse`        | Output the response of calling a function in the logs.                                                             | `true`                         |
| `printResponseBody`    | Output to the logs the response body when calling a function.                                                      | `false`                        |
| `printRequestBody`     | Output to the logs the request body when calling a function.                                                       | `false`                        |
| `fullnameOverride`     | Override the name value used for the Connector Deployment object.                                                  | ``                             |
| `contentType`          | Set a HTTP Content Type during function invocation.                                                                | `""`                           |
| `resources`            | Resources requests and limits configuration                                                                        | `requests.memory: "64Mi"`      |
| `logs.debug`           | Print debug logs                                                                                                   | `false`                        |
| `logs.format`          | The log encoding format. Supported values: `json` or `console`                                                     | `console`                      |

Specify each parameter using the `--set key=value[,key=value]` argument to `helm install`. See `values.yaml` for the default configuration.

## Removing the sns-connector

All control plane components can be cleaned up with helm:

```sh
$ helm uninstall -n openfaas sns-connector
```
