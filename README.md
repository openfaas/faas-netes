faas-netes - Serverless Functions For Kubernetes with OpenFaaS
===========

[![Build Status](https://github.com/openfaas/faas-netes/workflows/build/badge.svg?branch=master)](https://github.com/openfaas/faas-netes/actions)
[![Go Report Card](https://goreportcard.com/badge/github.com/openfaas/faas-netes)](https://goreportcard.com/report/github.com/openfaas/faas-netes)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![OpenFaaS](https://img.shields.io/badge/openfaas-serverless-blue.svg)](https://www.openfaas.com)

## Introduction

`faas-netes` is an [OpenFaaS provider](https://github.com/openfaas/faas-provider) which enables Kubernetes for [OpenFaaS](https://github.com/openfaas/faas). It's part of a larger stack that brings a cloud-agnostic serverless experience to Kubernetes.

The existing REST API, CLI and UI are fully compatible. With OpenFaaS Pro, you have an optional *operator* mode so that you can manage functions with `kubectl` and a `CustomResource`.

You can deploy OpenFaaS to any Kubernetes service - whether managed or local, including to OpenShift. You will find any specific instructions and additional links in the documentation.

[OpenFaaS (Functions as a Service)](https://github.com/openfaas/faas) is a framework for building serverless functions with Docker and Kubernetes which has first class support for metrics. Any process can be packaged as a function enabling you to consume a range of web events without repetitive boiler-plate coding.

<img alt="OpenFaaS workflow" src="https://raw.githubusercontent.com/openfaas/faas/master/docs/of-workflow.png"></img>
> Pictured: [OpenFaaS conceptual architecture](https://docs.openfaas.com/architecture/stack/)

## Highlights

* Platform for deploying [serverless-style workloads](https://docs.openfaas.com/reference/workloads/) - microservices and functions
* Native Kubernetes integrations (API and ecosystem)
* Built-in UI portal
* Scale to and from zero
* Built-in queuing and asynchronous invocations
* Custom routes and domain support
* Commercial support available

Additional & ecosystem:

* A range of event-connectors and cron-support
* helm chart and CLI installer
* Operator available to use Custom Resource Definitions (CRDs) (See also: [OpenFaaS Pro](https://openfaas.com/pricing))
* IDp integration with OIDC and commercial add-on

Community:

* Over 30k GitHub stars
* Independent open-source project with over 300 contributors, with commercial option available

Commercial options:

* Support from full-time team
* Commercial add-ons and integrations with events like Kafka, Postgres, AWS SQS and Cron
* Multiple namespace support
* gVisor support and runtimeClass for isolation
* Affinity and advanced scheduling / security constraints

## Get started

* Tutorial: [Deploy OpenFaaS to Kubernetes with its helm chart](https://docs.openfaas.com/deployment)
* [Read news and tutorials on the openfaas.com blog](https://www.openfaas.com/blog/)
* [Meet the community at the weekly Office Hours](https://docs.openfaas.com/community)

### The PLONK Stack

OpenFaaS can be used as complete stack for Cloud Native application development called PLONK. The PLONK Stack includes: Prometheus, Linux/Linkerd, OpenFaaS, NATS/Nginx and Kubernetes.

Read more: [Introducing PLONK](https://www.openfaas.com/blog/plonk-stack/).

## Technical and operational information

The rest of this document is dedicated to technical and operational information for the controller.

### Operating modes - classic or operator

There are two modes available for faas-netes, the classic mode is the default.

* Classic mode (aka faas-netes) - includes a REST API,  multiple-namespace support but no Function CRD - available in Community Edition and OpenFaaS Pro/Enteprise
* Operator mode (aka "The OpenFaaS Operator") - includes a REST API, with a "Function" CRD and multiple-namespace [OpenFaaS Pro/Enterprise](https://openfaas.com/pricing/)

See also: [README for "The OpenFaaS Operator"](README-OPERATOR.md)

The single faas-netes image and binary contains both modes, switch between one or the other using the helm chart or the flag `-operator=true/false`.

### Configuration of the controller

faas-netes can be configured with environment variables, but for a full set of options see the [helm chart](./chart/openfaas/).

| Option                      | Usage                                                                                            |
| --------------------------- | ------------------------------------------------------------------------------------------------ |
| `httpProbe`                 | Boolean - use http probe type for function readiness and liveness. Default: `true`              |
| `write_timeout`             | HTTP timeout for writing a response body from your function (in seconds). Default: `60s`         |
| `read_timeout`              | HTTP timeout for reading the payload from the client caller (in seconds). Default: `60s`         |
| `image_pull_policy`         | Image pull policy for deployed functions (`Always`, `IfNotPresent`, `Never`).  Default: `Always` |
| `gateway.resources`         | CPU/Memory resources requests/limits (memory: `120Mi`, cpu: `50m`)                               |
| `faasnetes.resources`       | CPU/Memory resources requests/limits (memory: `120Mi`, cpu: `50m`)                               |
| `operator.resources`        | CPU/Memory resources requests/limits (memory: `120Mi`, cpu: `50m`)                               |
| `queueWorker.resources`     | CPU/Memory resources requests/limits (memory: `120Mi`, cpu: `50m`)                               |
| `prometheus.resources`      | CPU/Memory resources requests/limits (memory: `512Mi`)                                           |
| `alertmanager.resources`    | CPU/Memory resources requests/limits (memory: `25Mi`)                                            |
| `nats.resources`            | CPU/Memory resources requests/limits (memory: `120Mi`)                                           |
| `faasIdler.resources`       | CPU/Memory resources requests/limits (memory: `64Mi`)                                            |
| `basicAuthPlugin.resources` | CPU/Memory resources requests/limits (memory: `50Mi`, cpu: `20m`)                                |

### Readiness checking

The readiness checking for functions assumes you are using our function watchdog which writes a .lock file in the default "tempdir" within a container. To see this in action you can delete the .lock file in a running Pod with `kubectl exec` and the function will be re-scheduled.

### Namespaces

By default all OpenFaaS functions and services are deployed to the `openfaas` and `openfaas-fn` namespaces. To alter the namespace use the `helm` chart.

### Ingress

To configure ingress see the `helm` chart. By default NodePorts are used. These are listed in the [deployment guide](https://docs.openfaas.com/deployment).

By default functions are exposed at `http://gateway:8080/function/NAME`.

You can also use the [IngressOperator to set up custom domains and HTTP paths](https://github.com/openfaas-incubator/ingress-operator)

### Image pull policy

By default, deployed functions will use an [imagePullPolicy](https://kubernetes.io/docs/concepts/containers/images/#updating-images) of `Always`, which ensures functions using static image tags are refreshed during an update.

If this is not desired behavior, OpenFaaS Pro customers can set the `image_pull_policy` environment variable to an alternative. `IfNotPresent` is particularly useful when developing locally with minikube. In this case, you can set your local environment to [use minikube's docker](https://kubernetes.io/docs/getting-started-guides/minikube/#reusing-the-docker-daemon) so `faas-cli build` builds directly into minikube's image store.
`faas-cli push` is unnecessary in this workflow - use `faas-cli build` then `faas-cli deploy`.

Note: When set to `Never`, **only** local (or pulled) images will work.  When set to `IfNotPresent`, function deployments may not be updated when using static image tags.

## Kubernetes Versions

faas-netes maintainers strive to support as many Kubernetes versions as possible and it is currently compatible with Kubernetes 1.11 and higher. Instructions for OpenShift are also available in the documentation.

## Contributing

You can quickly create a standard development environment using:

```sh
make start-kind
```

This will use [KinD](https://github.com/kubernetes-sigs/kind) to create a single node cluster and install the latest version of OpenFaaS via the Helm chart.

Check the contributor guide in `CONTRIBUTING.md` for more details on the workflow, processes, and additional tips.

## License

This project is licensed under the MIT License.
