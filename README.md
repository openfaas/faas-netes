faas-netes
===========

Poc for sandboxing a FaaS K8s integration

Create a cluster on Mac:

```
$ minikube start --vm-driver=xhyve
```

Start a "function"

```
$ kubectl run --labels="faas_function=true" --rm -i -t nodeinfo --port 8080 --image functions/nodeinfo:latest
```

Development:

```
$ eval $(minikube docker-env)
$ docker build -t faas-netesd .
$ kubectl delete service/faas-netesd ; \
   kubectl delete deployment/faas-netesd ; \
   kubectl run faas-netesd --image=faas-netesd:latest --port 8080 --image-pull-policy=Never ; \
   kubectl expose deployment/faas-netesd
```

Now try it out

This it the route for the function list as used by the FaaS UI / gateway.

```
$ kubectl get service faas-netesd
NAME          CLUSTER-IP   EXTERNAL-IP   PORT(S)    AGE
faas-netesd   10.0.0.46    <none>        8080/TCP   10s

$ minikube ssh 'curl -s 10.0.0.46:8080/system/functions'

[{"name":"nodeinfo","image":"functions/nodeinfo:latest","invocationCount":0,"replicas":1}]
```
