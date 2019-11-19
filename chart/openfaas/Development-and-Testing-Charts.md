# Local Chart Development and Testing

OpenFaaS Helm Charts can be developed and tested on the local system with ease.

## Requirements

Before you begin with your workflow make sure you have the following tools installed and operational.

* [Helm v2](https://github.com/helm/helm#install)
  * [Helm v2 Tiller plugin](https://github.com/rimusz/helm-tiller#installation)
  * [Helm Diff plugin](https://github.com/databus23/helm-diff#install)
* [Helmfile](https://github.com/roboll/helmfile/#installation)
* [Docker](https://docs.docker.com/install/)
* [k3d](https://github.com/rancher/k3d#get)

## Development Workflow

Before you start changing the charts verify that the current helm chart is working.

```sh
k3d create --name openfaas-test-helm #Create local k3s cluster in a container
export KUBECONFIG="$(k3d get-kubeconfig --name='openfaas-test-helm')" #Set kubeconfig location as environment variable
helmfile -f helmfile-local.yaml apply #Deploy current directory
helmfile -f helmfile-local.yaml test  #Test deployment directory
```

Once all test are working as expected you can move over to change the charts and start the **Inner Loop Workflow** to immediately test you results before the commit.

```sh
helmfile -f helmfile-local.yaml apply
helmfile -f helmfile-local.yaml test
```
