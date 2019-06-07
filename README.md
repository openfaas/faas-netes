faas-netes
===========

[![Go Report Card](https://goreportcard.com/badge/github.com/openfaas/faas-netes)](https://goreportcard.com/report/github.com/openfaas/faas-netes) [![Build Status](https://travis-ci.org/openfaas/faas-netes.svg?branch=master)](https://travis-ci.org/openfaas/faas-netes)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![OpenFaaS](https://img.shields.io/badge/openfaas-serverless-blue.svg)](https://www.openfaas.com)

`faas-netes` is an OpenFaaS provider which enables Kubernetes for [OpenFaaS](https://github.com/openfaas/faas). The existing REST API, CLI and UI are fully compatible.

> OpenFaaS also runs well on managed Kubernetes services like AKS, EKS and GKE. See our list of tutorials in the documentation site for more.

**Watch a video demo from [TechFieldDay Extra at Dockercon](https://www.youtube.com/watch?v=C3agSKv2s_w&list=PLlIapFDp305AiwA17mUNtgi5-u23eHm5j&index=1)**

[OpenFaaS (Functions as a Service)](https://github.com/openfaas/faas) is a framework for building serverless functions with Docker and Kubernetes which has first class support for metrics. Any process can be packaged as a function enabling you to consume a range of web events without repetitive boiler-plate coding.

![Stack](https://camo.githubusercontent.com/08bc7c0c4f882ef5eadaed797388b27b1a3ca056/68747470733a2f2f7062732e7477696d672e636f6d2f6d656469612f4446726b46344e586f41414a774e322e6a7067)

In this README you'll find a technical overview and instructions for deploying to a Kubernetes cluster. (Docker Swarm is also supported along with Hashicorp Nomad)

* Framework for deploying serverless workloads
* Native Kubernetes integrations (API and ecosystem)
* Built-in UI
* YAML templates & helm chart
* Over 13.5k GitHub stars
* Independent open-source project with over 130 contributors
* Operator available to use Custom Resource Definitions (CRDs) [openfaas-operator](https://github.com/openfaas-incubator/openfaas-operator/)

## Get started

* Tutorial: [Deploy on Kubernetes with helm/YAML](https://docs.openfaas.com/deployment)
* [Visit the website](https://www.openfaas.com)
* [Join the community](https://docs.openfaas.com/community)

### How is this project different from others?

* [Read an Introduction to OpenFaaS here](https://blog.alexellis.io/introducing-functions-as-a-service/)

## Reference guide

### Configuration via Environmental variables

faas-netes can be configured with environment variables, but for a full set of options see the [helm chart](./chart/openfaas/).

| Option              | Usage                                                                                           |
|---------------------|-------------------------------------------------------------------------------------------------|
| `httpProbe`         | Boolean - use http probe type for function readiness and liveness. Default: `false`             |
| `write_timeout`     | HTTP timeout for writing a response body from your function (in seconds). Default: `60s`        |
| `read_timeout`      | HTTP timeout for reading the payload from the client caller (in seconds). Default: `60s`        |
| `image_pull_policy` | Image pull policy for deployed functions (`Always`, `IfNotPresent`, `Never`.  Default: `Always` |

### Readiness checking

The readiness checking for functions assumes you are using our function watchdog which writes a .lock file in the default "tempdir" within a container. To see this in action you can delete the .lock file in a running Pod with `kubectl exec` and the function will be re-scheduled.

### Namespaces

By default all OpenFaaS functions and services are deployed to the `openfaas` and `openfaas-fn` namespaces. To alter the namespace use the `helm` chart.

### Ingress

To configure ingress see the `helm` chart. By default NodePorts are used. These are listed in the [deployment guide](https://docs.openfaas.com/deployment).

### Image pull policy

By default, deployed functions will use an [imagePullPolicy](https://kubernetes.io/docs/concepts/containers/images/#updating-images) of `Always`, which ensures functions using static image tags are refreshed during an update.
If this is not desired behavior, set the `image_pull_policy` environment variable to an alternative.  `IfNotPresent` is particularly useful when developing locally with minikube.
In this case, you can set your local environment to [use minikube's docker](https://kubernetes.io/docs/getting-started-guides/minikube/#reusing-the-docker-daemon) so `faas-cli build` builds directly into minikube's image store.
`faas-cli push` is unnecessary in this workflow - use `faas-cli build` then `faas-cli deploy`.

Note: When set to `Never`, **only** local (or pulled) images will work.  When set to `IfNotPresent`, function deployments may not be updated when using static image tags.

## Kubernetes Versions
OpenFaaS strives to support as many Kubernetes versions as possible. Due to recent changes in the Secrets API, **Kubernetes 1.9** or higher is required with **faas-netes 0.7.1** or higher.

## Contributing to faas-netes

You can quickly create a standard development environment using

```sh
make start-kind
```

this will use [KinD](https://github.com/kubernetes-sigs/kind) to create a single node cluster and install the latest version of OpenFaaS via the Helm chart.

Check the contributor guide in `CONTRIBUTING.md` for more details on the workflow, processes, and additional tips.
