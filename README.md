faas-netes
===========

[![Go Report Card](https://goreportcard.com/badge/github.com/openfaas/faas-netes)](https://goreportcard.com/report/github.com/openfaas/faas-netes) [![Build Status](https://travis-ci.org/openfaas/faas-netes.svg?branch=master)](https://travis-ci.org/openfaas/faas-netes)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![OpenFaaS](https://img.shields.io/badge/openfaas-serverless-blue.svg)](https://www.openfaas.com)

This is a plugin to enable Kubernetes as an [OpenFaaS](https://github.com/openfaas/faas) backend. The existing CLI and UI are fully compatible. It also opens up the possibility for other plugins to be built for orchestation frameworks such as Nomad,  Mesos/Marathon or even a cloud-managed back-end such as Hyper.sh or Azure ACI.

**Update:** [Watch the demo and intro to the CNCF Serverless Workgroup](https://youtu.be/SwRjPiqpFTk?t=1m8s)

[OpenFaaS](https://github.com/openfaas/faas) is an event-driven serverless framework for containers. Any container for Windows or Linux can be leveraged as a serverless function. OpenFaaS is quick and easy to deploy (less than 60 secs) and lets you avoid writing boiler-plate code.

![Stack](https://camo.githubusercontent.com/08bc7c0c4f882ef5eadaed797388b27b1a3ca056/68747470733a2f2f7062732e7477696d672e636f6d2f6d656469612f4446726b46344e586f41414a774e322e6a7067)

In this README you'll find a technical overview and instructions for deploying FaaS on a Kubernetes cluster. 

> Docker Swarm is also supported. 

* Serverless framework for containers
* Native Kubernetes support
* Built-in UI
* YAML templates
* Over 6.5k GitHub stars
* Independent open-source project with 35 authors/contributors

You can watch my intro from [the Dockercon closing keynote with Alexa, Twitter and Github demos](https://www.youtube.com/watch?v=-h2VTE9WnZs&t=910s) or a [complete walk-through of FaaS-netes](https://www.youtube.com/watch?v=0DbrLsUvaso) showing Prometheus, auto-scaling, the UI and CLI in action.

If you'd like to know more about the OpenFaaS project head over to - https://github.com/openfaas/faas

## Get started

If you're looking to just get OpenFaaS deployed on Kubernetes follow the [OpenFaaS and Kubernetes Deployment Guide](https://github.com/openfaas/faas/blob/master/guide/deployment_k8s.md) or read on for a technical overview.

To try our `helm` chart click here: [helm guide](https://github.com/openfaas/faas-netes/blob/master/HELM.md).

### How is this project different from others?

* [Read an Introduction to OpenFaaS here](https://blog.alexellis.io/introducing-functions-as-a-service/).

## Reference guide

### Configuration

FaaS-netes can be configured via environment variables.

**Environmental variables:**

| Option                 | Usage                                                                                          |
|------------------------|------------------------------------------------------------------------------------------------|
| `enable_function_readiness_probe` | Boolean - enable a readiness probe to test functions. Default: `true`               |
| `write_timeout`        | HTTP timeout for writing a response body from your function (in seconds). Default: `8`         |
| `read_timeout`         | HTTP timeout for reading the payload from the client caller (in seconds). Default: `8`         |

**Readiness checking**

The readiness checking for functions assumes you are using our function watchdog which writes a .lock file in the default "tempdir" within a container. To see this in action you can delete the .lock file in a running Pod with `kubectl exec` and the function will be re-scheduled.

**Namespaces**

By default all OpenFaaS functions and services are deployed to the `default` namespace. To use a separate namespace for functions and the services use the Helm chart.

**Asynchronous processing**

To enable asynchronous processing use the helm chart configuration.

### Technical overview

The code in this repository is a daemon or micro-service which can provide the basic functionality the FaaS Gateway requires:

* List functions
* Deploy function
* Delete function
* Invoke function synchronously

Any other metrics or UI components will be maintained separately in the main OpenFaaS project.

**Motivation for separate micro-service:**

* Kubernetes go-client is 41MB with only a few lines of code
* After including the go-client the code takes > 2mins to compile

So rather than inflating the original project's source-code this micro-service will act as a co-operator or plug-in. Some additional changes will be needed in the main OpenFaaS project to switch between implementations.

![](https://pbs.twimg.com/media/DFh7i-ZXkAAZkw4.jpg:large)

There is no planned support for dual orchestrators - i.e. Swarm and K8s at the same time on the same host/network.

### Get involved

*Please Star the FaaS and FaaS-netes Github repo.*

* [Main OpenFaaS repo](https://github.com/openfaas/faas)

Contributions are welcome - see the [contributing guide for OpenFaaS](https://github.com/openfaas/faas/blob/master/CONTRIBUTING.md).

The [OpenFaaS complete walk-through on Kubernetes Video](https://www.youtube.com/watch?v=0DbrLsUvaso) shows how to use Prometheus and the auto-scaling in action.

### Explore OpenFaaS / FaaS-netes with minikube

> These instructions may be out of sync with the latest changes. If you're looking to just get OpenFaaS deployed on Kubernetes follow the [OpenFaaS and Kubernetes Deployment Guide](https://github.com/openfaas/faas/blob/master/guide/deployment_k8s.md).

Let's try it out:

* Create a single-node cluster on our Mac
* Deploy a function with the `faas-cli`
* Deploy OpenFaaS with `helm`
* Make calls to list the functions and invoke a function

I'll give instructions for creating your cluster on a Mac with `minikube`, but you can also use `kubeadm` on Linux in the cloud by [following this tutorial](https://blog.alexellis.io/kubernetes-kubeadm-video/).

[Begin the Official tutorial on Medium](https://medium.com/@alexellisuk/getting-started-with-openfaas-on-minikube-634502c7acdf)

### Appendix

**Auto-scale your functions**

Given enough load (> 5 requests/second) OpenFaaS will auto-scale your service, you can test this out by opening up the Prometheus web-page and then generating load with Apache Bench or a while/true/curl bash loop.

Here's an example you can use to generate load:

```
ip=$(minikube ip); while [ true ] ; do curl $ip:31112/function/nodeinfo -d "" ; done
```

Prometheus is exposed on a NodePort of 31119 which shows the function invocation rate. 

```
$ open http://$(minikube ip):31119/
```

Here's an example for use with the Prometheus UI:

```
rate(gateway_function_invocation_total[20s])
```

> It shows the rate the function has been invoked over a 20 second window.

The [OpenFaaS complete walk-through on Kubernetes Video](https://www.youtube.com/watch?v=0DbrLsUvaso) shows auto-scaling in action and how to use the Prometheus UI.

**Test out the UI**

You can also access the OpenFaaS UI through the node's IP address and the NodePort we exposed earlier.

```
$ open http://$(minikube ip):31112/
```

![](https://pbs.twimg.com/media/DFkUuH1XsAAtNJ6.jpg:medium)

If you've ever used the *Kubernetes dashboard* then this UI is a similar concept. You can list, invoke and create new functions.
