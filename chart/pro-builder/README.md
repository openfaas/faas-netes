# OpenFaaS Pro Function Builder Chart

Build OpenFaaS functions via a REST API

The [pro-builder](https://docs.openfaas.com/openfaas-pro/builder/) is used to build functions from within your cluster using a REST API.

## Prerequisites

- Obtain a license key

  You will need a license for OpenFaaS Pro, contact us at: [openfaas.com/pricing](https://www.openfaas.com/pricing)

- A container image registry that is accessible from your cluster

  You can generate a valid container registry login file by running `faas-cli registry-login`, or by disabling the keychain in Docker, running `docker login` and then using your own `$HOME/.docker/config.json` file.

- OpenFaaS pre-installed

  You do not need OpenFaaS to be installed to use the builder, however you will need the openfaas namespace to have been created.

## Installation

Create a registry push secret for the Pro Builder to use to push images to your registry. This can be generated through `docker login`.

```bash
kubectl create secret generic registry-secret \
    --from-file config.json=$HOME/.docker/config.json -n openfaas
```

> For pushing images to ECR see: [Push images to Amazon ECR](#push-images-to-amazon-ecr)

Create a HMAC signing secret for use between the Pro Builder and your client:

```bash
PAYLOAD=$(head -c 32 /dev/urandom |shasum | cut -d ' ' -f 1)

# Save a copy for later use
echo $PAYLOAD > payload.txt
```

Create a secret with the contents of the signing secret:

```bash
kubectl create secret generic payload-secret \
  --from-file payload-secret=payload.txt -n openfaas
```

Generate mTLS certificates for BuildKit and the Pro Builder which are used to encrypt messages between the builder component and BuildKit.

```bash
docker run -v `pwd`/out:/tmp/ -ti ghcr.io/openfaas/certgen:latest

# Reset the permissions of the files to your own user:
sudo chown -R $USER:$USER out
```

Then create two secrets, one for the BuildKit daemon and one for the builder component.

```bash
kubectl create secret generic -n openfaas \
  buildkit-daemon-certs \
  --from-file ./out/certs/ca.crt \
  --from-file ./out/certs/server.crt \
  --from-file ./out/certs/server.key

kubectl create secret generic -n openfaas \
  buildkit-client-certs \
  --from-file ./out/certs/ca.crt \
  --from-file ./out/certs/client.crt \
  --from-file ./out/certs/client.key
```

## Install the Chart

- Create the required secret with your OpenFaaS Pro license code:

```bash
$ kubectl create secret generic \
    -n openfaas \
    openfaas-license \
    --from-file license=$HOME/.openfaas/LICENSE
```

- Add the OpenFaaS chart repo and deploy the `pro-builder` Pro chart. We recommend installing it in the same namespace as the rest of OpenFaaS

```sh
$ helm repo add openfaas https://openfaas.github.io/faas-netes/
$ helm repo update
```

Next, you need to decide whether you're going to run the builder as root, or in a rootless, restricted mode for production.

Create a `custom-values.yaml` file accordingly.

Root mode, for development, or where rootless for some reason isn't working:

```yaml
buildkit:
  image: moby/buildkit:v0.10.3
  rootless: false
  securityContext:
    runAsUser: 0
    runAsGroup: 0
    privileged: true
```

Rootless mode (preferred, if possible):

```yaml
buildkit:
  # For a rootless configuration
  image: moby/buildkit:master-rootless
  rootless: true
  securityContext:
    # Needs Kubernetes >= 1.19
    seccompProfile:
      type: Unconfined
    runAsUser: 1000
    runAsGroup: 1000
    privileged: false
```

Then install the chart using its official path and the custom YAML file:

```sh
$ helm upgrade pro-builder openfaas/pro-builder \
    --install \
    --namespace openfaas \
    -f custom-values.yaml
```

When installing from the faas-netes repository on your local computer, whilst making changes or customising the chart, you can run this command instead from the faas-netes folder:

```sh
$ helm upgrade pro-builder ./chart/pro-builder \
    --install \
    --namespace openfaas \
    --values ./chart/pro-builder/values.yaml
```

The Pod for the builder contains two containers:

* pro-builder - the REST API for building functions
* buildkit - the daemon for Buildkit

Pass either to the logs command:

```
# Check the logs of the pro-builder API
kubectl logs -n openfaas \
  deploy/pro-builder -c pro-builder

# Check the logs of the buildkit daemon
kubectl logs -n openfaas \
  deploy/pro-builder -c buildkit

# Check the events for any missing secrets or volumes
kubectl get events -n openfaas
```

To test the builder head over to the [Function Builder API Documentation](https://docs.openfaas.com/openfaas-pro/builder/)

## Troubleshooting

### Errors due to permissions

If you see errors about permissions, then you may need to review the options for the securityContext.

See also: [rootless mode](https://github.com/moby/buildkit/blob/master/docs/rootless.md)

### Errors due to authentication

If you're having issues getting your registry authentication to work, then why not try out ttl.sh, a free, ephemeral container registry. [ttl.sh](https://ttl.sh) is a public service run by Replicated, which allows you to push and pull images without authentication.

Once you've seen the building work end to end, get in touch with us and we'll try to help you with your authentication.

## Push images to Amazon ECR

The pro-builder can be configured to push images to Amazon ECR.

Configure the Pro Builder to use the Amazon ECR credential helper. The required `config.json` file can be generated by the `faas-cli`:

```bash
faas-cli registry-login --ecr \
  --acount-id <aws_account_id> \
  --region <region>
```

```
kubectl create secret generic registry-secret \
    --from-file config.json=credentials/config.json -n openfaas
```

> For more details on the configuration see: [amazon-ecr-credentials-helper](https://github.com/awslabs/amazon-ecr-credential-helper#docker)

Create a secret for the AWS credentials. The credentials must have a policy applied that  allows access to Amazon ECR.

```
kubectl create secret generic -n openfaas \
  aws-ecr-credentials --from-file aws-ecr-credentials=$HOME/ecr-credentials.txt
```

Modify your `values.yaml` file accordingly to mount the secret in the Pro Builder.

```yaml
awsCredentialsSecret: aws-ecr-credentials
```

> Keep in mind that ECR requires a repository to be pre created before any image can be pushed.

## Configuration

Additional pro-builder options in `values.yaml`.

| Parameter                 | Description                                                                            | Default                        |
| ------------------------- | -------------------------------------------------------------------------------------- | ------------------------------ |
| `replicas`                | How many replicas of buildkit and the pro-builder API to create                        | `1`                            |
| `proBuilder.image`        | Container image to use for the pro-builder                                             | See values.yaml                |
| `proBuilder.maxInflight`  | Limit the total amount of concurrent builds for the  pro-builder replica               | See values.yaml                |
| `buildkit.image`          | Image version for the buildkit daemon                                                  | See values.yaml                |
| `buildkit.rootless`       | When set to true, uses user-namespaces to avoid a privileged container                 | See notes in values.yaml       |
| `buildkit.securityContext` | Used to set security policy for buildkit                                              | See values.yaml                |
| `imagePullPolicy`         | The policy for pulling either of the containers deployed by this chart                 | `IfNotPresent`                 |
| `disableHmac`             | This setting disable request verification, so should never to set to `true`            | `false`                        |
| `enableLchown`            | Toggle whether Lchown is used by buildkit, toggle if you run into errors               | `false`                        |
| `awsCredentialsSecret` | Mount a secret with AWS credentials for pushing images to ECR | `""` |

Specify each parameter using the `--set key=value[,key=value]` argument to `helm install`. See `values.yaml` for the default configuration.

## Removing the pro-builder

All control plane components can be cleaned up with helm:

```bash
$ helm uninstall pro-builder \
  --namespace openfaas

$ kubectl delete -n openfaas \
    secret/registry-secret

$ kubectl delete -n openfaas \
    secret/payload-secret

$ kubectl delete -n openfaas \
  secret/buildkit-daemon-certs

$ kubectl delete -n openfaas \
  secret/buildkit-client-certs
```
