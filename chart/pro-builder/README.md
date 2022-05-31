# OpenFaaS Pro Function Builder Chart

Build OpenFaaS functions via a REST API

The [pro-builder](https://docs.openfaas.com/openfaas-pro/builder/) is used to build functions from within your cluster using a REST API.

## Prerequisites

- Obtain a license or trial

  You will need a license for OpenFaaS Pro, contact us at: [openfaas.com/support](https://www.openfaas.com/support)

- Install OpenFaaS

  You must have a working OpenFaaS installed.

## Installation

Create a registry push secret for the Pro Builder to use to push images to your registry. This can be generated through `docker login`.

```bash
kubectl create secret generic registry-secret \
    --from-file config.json=$HOME/.docker/config.json -n openfaas
```

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

Generate mTLS certificates for BuildKit and the Pro Builder which are used to encrypt messages between the builder component and buildkit.

```bash
docker run -v `pwd`/out:/tmp/ -ti ghcr.io/openfaas/certgen:0.1.0-rc2

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
$ helm upgrade pro-builder openfaas/pro-builder \
    --install \
    --namespace openfaas
```

When installing from the faas-netes repository on your local computer:

```sh
$ helm upgrade pro-builder ./chart/pro-builder \
    --install \
    --namespace openfaas \
    --values ./chart/pro-builder/values.yaml
```

> The above command will also update your helm repo to pull in any new releases.

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

To test the builder head over to the [OpenFaaS Documentation](https://docs.openfaas.com/openfaas-pro/builder/)

## Troubleshooting

If you see errors about permissions, then you may need to review the options for the securityContext.

See also: [rootless mode](https://github.com/moby/buildkit/blob/master/docs/rootless.md)

## Configuration

Additional pro-builder options in `values.yaml`.

| Parameter                | Description                                                                            | Default                        |
| ------------------------ | -------------------------------------------------------------------------------------- | ------------------------------ |
| `image`                  | Container image to use for the pro-builder            | See values.yaml                 |

Specify each parameter using the `--set key=value[,key=value]` argument to `helm install`. See `values.yaml` for the default configuration.

## Removing the pro-builder

All control plane components can be cleaned up with helm:

```sh
$ helm uninstall pro-builder --namespace openfaas
```
