# kail: kubernetes tail [![Build Status](https://travis-ci.org/boz/kail.svg?branch=master)](https://travis-ci.org/boz/kail)

Kubernetes tail.  Streams logs from all containers of all matched pods.  Match pods by service, replicaset, deployment, and others.  Adjusts to a changing cluster - pods are added and removed from logging as they fall in or out of the selection.

[![asciicast](https://asciinema.org/a/133521.png)](https://asciinema.org/a/133521)

## Usage

With no arguments, kail matches all pods in the cluster.  You can control the matching pods with arguments which select pods based on various criteria.

### Selectors

Flag | Selection
--- | ---
`-l, --label LABEL-SELECTOR` | match pods based on a [standard label selector](https://kubernetes.io/docs/concepts/overview/working-with-objects/labels/)
`-p, --pod NAME` | match pods by name
`-n, --ns NAMESPACE-NAME` | match pods in the given namespace
`--svc NAME` | match pods belonging to the given service
`--rc NAME` | match pods belonging to the given replication controller
`--rs NAME` | match pods belonging to the given replica set
`-d, --deploy NAME` | match pods belonging to the given deployment
`--node NODE-NAME` | match pods running on the given node
`--ing NAME` | match pods belonging to services targeted by the given ingress
`-c, --containers CONTAINER-NAME` | restrict which containers logs are shown for
`--ignore LABEL-SELECTOR` | Ignore pods that the selector matches. (default: `kail.ignore=true`)

#### Name Selection

When selecting objects by `NAME` (`--svc`, `--pod`, etc...), you can either qualify it with a namespace to restrict the selection to the given namespace, or select across all namespaces by giving just the object name.

Example:

```sh
# match pods belonging to a replicaset named 'workers' in any namespace.
$ kail --rs workers

# match pods belonging to the replicaset named 'workers' only in the 'staging' namespace
$ kail --rs staging/workers
```

#### Combining Selectors

If the same flag is used more than once, the selectors for that flag are "OR"ed together.

```sh
# match pods belonging to a replicaset named "workers" or "db"
$ kail --rs workers --rs db
```

Different flags are "AND"ed together:

```sh
# match pods belonging to both the service "frontend" and the deployment "webapp"
$ kail --svc frontend --deploy webapp
```

### Other Flags

Flag | Description
--- | ---
`-h, --help` | Display help and usage
`--context CONTEXT-NAME` | Use the given Kubernetes context
`--dry-run` | Print initial matched pods and exit
`--log-level LEVEL` | Set the logging level (default: `error`)
`--log-file PATH` | Write output to `PATH` (default: `/dev/stderr`)
`--since DURATION` | Display logs as old as given duration. Ex: `5s`, `2m`, `1.5h` or `2h45m` (defaults: `1s`)
See [here](https://golang.org/pkg/time/#ParseDuration) for more information on the duration format.

## Installing

### Homebrew

```sh
$ brew tap boz/repo
$ brew install boz/repo/kail
```

### Downloading

Kail binaries for Linux and OSX can be found on the [latest release](https://github.com/boz/kail/releases/latest) page.  Download and install into your `$GOPATH/bin` with:

```sh
$ bash <( curl -sfL https://raw.githubusercontent.com/boz/kail/master/godownloader.sh) -b "$GOPATH/bin"
```

### Running in a cluster with `kubectl`

The docker image [abozanich/kail](https://hub.docker.com/r/abozanich/kail/) is available for running `kail` from within a kubernetes pod via `kubectl`.

Note: be sure to include the `kail.ignore=true` label, otherwise... it's logging all the way down.

Example:

```sh
# match all pods - synonymous with 'kail' from the command line
$ kubectl run -it --rm -l kail.ignore=true --restart=Never --image=abozanich/kail kail

# match pods belonging to service 'api' in any namespace - synonymous with 'kail --svc api'
$ kubectl run -it --rm -l kail.ignore=true --restart=Never --image=abozanich/kail kail -- --svc api
```

## Building

### Install build and dev dependencies

* [govendor](https://github.com/kardianos/govendor)
* [minikube](https://kubernetes.io/docs/getting-started-guides/minikube/)
* _linux only_: [musl-gcc](https://www.musl-libc.org/how.html) for building docker images.

### Install source code and golang dependencies

```sh
$ go get -d github.com/boz/kail
$ cd $GOPATH/src/github.com/boz/kail
$ make install-deps
```

### Build binary

```sh
$ make
```

### Install run against a demo cluster

```sh
$ minikube start
$ ./_example/demo.sh start
$ ./kail

# install image into minikube and run via kubectl
$ make image-minikube
$ kubectl run -it --rm -l kail.ignore=true --restart=Never --image=kail kail
```

### Stress testing

Start minikube with a fair amount of CPU and memory.

```sh
$ minikube start --cpus 4 --memory 8192
```

Start [`stress.yml`](_example/stress.yml) resources.

```sh
$ kubectl create -f _example/stress.yml
```

Wait a while for the pods to run

```sh
$ kubectl get pods --namespace stress | grep Running | wc -l
100
```

Run kail

```sh
./kail --ns stress
```
