## Contributing

### License

This project is licensed under the MIT License.

## Updating versions

### Helm

Any changes to the helm chart will also need a corresponding change in the [Chart.yml](https://github.com/openfaas/faas-netes/blob/master/chart/openfaas/Chart.yaml) file to bump up the version.
For version updates please review the [Semantic Versioning](https://semver.org/spec/v0.1.0.html) guidelines.

### ARM builds

ARM builds are provided on a best effort basis by Alex Ellis. If you need an updated ARM build you can build your own or contact Alex via Twitter/Slack to request an updated official image.

## Guidelines

See guide for [FaaS](https://github.com/openfaas/faas/blob/master/CONTRIBUTING.md) here.


## Getting Started

You can quickly create a standard development environment using

```sh
make start-kind
```

this will use [KinD](https://github.com/kubernetes-sigs/kind) to create a single node cluster and install the latest version of OpenFaaS via the Helm chart.

This environment has two goals:

1. match the environment we use in CI for integration tests and validating the Helm chart. We use the same scripts in CI to setup and test function deployments in faas-netes.
2. as much as is possible, enable completely offline and cost-free development, i.e. Airplane mode. If you have the latest base images for Go, you should not require an internet connection.



### Customizing the environment name
You can use `OF_DEV_ENV` to set a custom name for the cluster. The default value is `kind`.

### Loading your own images
As you are developing on `faas-netes`, you will want to build and test your own local images.  This can easily be done using the following commands

```sh
make build
kind load docker-image --name="${OF_DEV_ENV:-kind}" openfaas/faas-netes:latest
export KUBECONFIG="$(kind get kubeconfig-path --name="${OF_DEV_ENV:-kind}")"
helm upgrade openfaas --install openfaas/openfaas \
    --namespace openfaas  \
    --set basic_auth=true \
    --set openfaasImagePullPolicy=IfNotPresent \
    --set faasnetes.imagePullPolicy=IfNotPresent \
    --set faasnetes.image=openfaas/faas-netes:latest \
    --set functionNamespace=openfaas-fn
```

Note that this technique can also be used to test locally built functions, thus avoiding the need to push the image to a remote registry.

### Port Forwarding
It will port-forward the Gateway to your local 31112.  You can stop it using

```sh
kill $(<"of_${OF_DEV_ENV:-kind}_portforward.pid")
```

To manually restart the port-forward, use

```sh
kubectl port-forward deploy/gateway -n openfaas 31112:8080 &>/dev/null & echo -n "$!" > "of_${OF_DEV_ENV:-kind}_portforward.pid"
```

For simplicity, a restart script is provided in `contrib`

```sh
./contrib/restart_port_forward.sh
```

### Stopping the environment
Or stop the entire environment and cleanup using

```sh
make stop-kind
```
