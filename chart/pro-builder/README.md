# OpenFaaS Pro Function Builder Chart

Build OpenFaaS functions via a REST API

The [pro-builder](https://docs.openfaas.com/openfaas-pro/builder/) is used to build functions from within your cluster using a REST API.

## Prerequisites

- Obtain a license key

  You will need a license for OpenFaaS Pro, contact us at: [openfaas.com/pricing](https://www.openfaas.com/pricing)

- A container image registry that is accessible from your cluster

  You can generate a valid container registry login file by:
  
  * Running `faas-cli registry-login` (preferred)
  * Or, disable the keychain in Docker, then run `docker login`, and supply the `$HOME/.docker/config.json` file.

- OpenFaaS pre-installed

  You do not need OpenFaaS to be installed to use the builder, however you will need the openfaas namespace to have been created.

## Installation

Create a registry push secret for the Pro Builder to use to push images to your registry.

### Registry authentication

For testing with ttl.sh, create an empty auths section:

```bash
cat << EOF > ttlsh-config.json
{
  "auths": {}
}
EOF
```

Then create the secret within the cluster:

```bash
kubectl create secret generic registry-secret \
    --from-file config.json=./ttlsh-config.json -n openfaas
```

When you want to authenticate to a private registry, you must either:

* Use `faas-cli registry-login`, to create a file in the `.credentials` folder of the current working directory
* Or, turn off the credential store for Docker Desktop, delete `~/docker/config.json` and then run: `docker login` and enter your credentials.

And remember to delete any existing secret from the cluster first: `kubectl delete secret registry-secret -n openfaas`.

For pushing images to ECR see: [Push images to Amazon ECR](#push-images-to-amazon-ecr)

### Signing secret

Create a HMAC signing secret for use between the Pro Builder and your client:

```bash
echo -n $(openssl rand -base64 32) > payload.txt
```

Create a secret with the contents of the signing secret:

```bash
kubectl create secret generic payload-secret \
  --from-file payload-secret=payload.txt -n openfaas
```

### mTLS certificates

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

```bash
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

## Push images to Amazon ECR

Your image name should be of a format similar to: `ACCOUNT.dkr.ecr.REGION.amazonaws.com/REPOSITORY/IMAGE_NAME:IMAGE_TAG`

Bear in mind that unlike i.e. the Docker Hub or ghcr.io, Amazon ECR requires all repositories to be created via a [CreateRepository API call](https://docs.aws.amazon.com/AmazonECR/latest/APIReference/API_CreateRepository.html) before you can publish images to them from a container builder.

This can be achieved simply using a function or some custom code using one of the various [AWS SDKs](https://aws.amazon.com/developer/tools/).

During development and testing, you can create a repository manually using the AWS Console or CLI.

You can create a function that's quick to build with something like this command that users the dockerfile template:

```bash
export ACCOUNT="" # Your AWS account ID
export REGION="us-east-1" # Your AWS region
export ECR_REPO="of-test"
export IMAGE_NAME="fn"

faas-cli new --lang dockerfile \
  --prefix $ACCOUNT.dkr.ecr.$REGION/ $IMAGE_NAME
```

### Push using a shared credential (recommended for initial testing)

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

Create a secret for the AWS credentials file. The credentials must have a policy applied that allows access to Amazon ECR.

Create the credentials file:
```bash
export AWS_ACCESS_KEY_ID=""
export AWS_ACCESS_SECRET_KEY=""

cat > ecr-credentials.txt <<EOF
[default]
aws_access_key_id=$AWS_ACCESS_KEY_ID
aws_secret_access_key=$AWS_ACCESS_SECRET_KEY
EOF
```

```
kubectl create secret generic -n openfaas \
  aws-ecr-credentials --from-file aws-ecr-credentials=./ecr-credentials.txt
```

Modify your `values.yaml` file accordingly to mount the secret in the Pro Builder.

```yaml
awsCredentialsSecret: aws-ecr-credentials
```

### Push using AWS IAM roles and IAM roles for service accounts (IRSA)

If you are running on EKS, you can use IAM roles for service accounts (IRSA) to allow the Pro Builder to push images to ECR without creating a shared credential file.

You must first configure IRSA properly in your EKS cluster, which may involve configuring add-ons and an OIDC provider in your AWS account. For more instructions see the AWS documentation on [IAM roles for service accounts](https://docs.aws.amazon.com/eks/latest/userguide/iam-roles-for-service-accounts.html).

Create an appropriate policy for the pro-builder Pod to use with the ECR credential helper.

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "ecr:GetAuthorizationToken",
        "ecr:BatchCheckLayerAvailability",
        "ecr:GetDownloadUrlForLayer",
        "ecr:GetRepositoryPolicy",
        "ecr:DescribeRepositories",
        "ecr:ListImages",
        "ecr:DescribeImages",
        "ecr:BatchGetImage",
        "ecr:InitiateLayerUpload",
        "ecr:UploadLayerPart",
        "ecr:CompleteLayerUpload",
        "ecr:PutImage"
      ],
      "Resource": "*"
    }
  ]
}
```

Attach it to a service account named `pro-builder-ecr` in the `openfaas` namespace.

Next, when you create the `registry-secret`, configure the credential helper to use the ECR credential helper.

Update the ACCOUNT and REGION appropriately from the following example:

```json
{
  "credHelpers": {
    "ACCOUNT.dkr.ecr.REGION.amazonaws.com": "ecr-login"
  }
}
```

Edit values.yaml, ensure you have the following in place:

```yaml
serviceAccount: pro-builder-ecr

securityContext:
  fsGroup: 65534
```

Then install the chart using Helm.

You can verify that the service account is properly configured through an exec command.

```bash
kubectl exec --namespace openfaas -t -i  deploy/pro-builder -c pro-builder -- cat /var/run/secrets/kubernetes.io/serviceaccount/token ; echo
```

Copy the output to [https://jwt.io](https://jwt.io)

* The `sub` field should be `"sub": "system:serviceaccount:openfaas:pro-builder-ecr"`
* The `iss` field should be `"iss": "https://oidc.eks.REGION.amazonaws.com/id/ACCOUNT_ID"`
* The `name` field under `serviceaccount` under `kubernetes.io` should read: `"pro-builder-ecr"`

For further testing, run the following:

```bash
kubectl exec --namespace openfaas -t -i  deploy/pro-builder -c pro-builder -- docker-credential-ecr-login list; echo
```

This command parses the registry secret to check the helper is configured properly, you should see a repository such as:

```
{"https://ACCOUNT.dkr.ecr.REGION.amazonaws.com":"AWS"}
```

In our testing, whilst the pro-builder does push images properly, the `docker-credential-ecr-login get` command appears to hang. Images are still pushed successfully, so this is not an issue.

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


## Troubleshooting

### Pending Pods

Check for events in the `openfaas` namespace, this may reveal issues with missing secrets, or too few resources available to schedule the Pod.

```bash
$ kubectl get events -n openfaas -o wide --watch
```

Also, verify if the nodes have enough resources available to run the pro-builder and buildkit containers. You can check this with:

```bash
$ kubectl describe nodes
```

If needed, reduce the resource requests in `values.yaml` for the `buildkit` and `pro-builder` containers, or add more nodes to your cluster.

### Errors due to permissions

If you see errors about permissions, then you may need to review the options for the securityContext.

See also: [rootless mode](https://github.com/moby/buildkit/blob/master/docs/rootless.md)

### Errors due to authentication

If you're having issues getting your registry authentication to work, then why not try out ttl.sh, a free, ephemeral container registry. [ttl.sh](https://ttl.sh) is a public service run by Replicated, which allows you to push and pull images without authentication.

Once you've seen the building work end to end, get in touch with us and we'll try to help you with your authentication.

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
