# Using OpenFaaS on OpenShift

The following describes the setup using [Minishift](https://github.com/minishift/minishift).

## Minishift setup

The setup has been tested with Minishift in version `v1.10.0+10461c6` using the following configuration:

```
$ minishift start --vm-driver=virtualbox --cpus=6 --memory=4GB --logging --metrics
$ oc login -u system:admin
$ oc policy add-role-to-user cluster-admin developer
$ oc policy add-role-to-user admin developer
$ oc login `minishift console --url` -u developer -p developer
```

## OpenFaaS setup

Get stuff from https://github.com/mhausenblas/faas-netes with a PR in the `openshift` branch in [upstream](https://github.com/openfaas/faas-netes) pending.

### Deployment

To set up OpenFaaS in OpenShift, we first need to create two projects (that's OpenShift talk for namespaces on steroids) in this order:

```
$ oc new-project openfaas-fn
$ oc new-project openfaas
```

Next we create the `faas-controller` service account and set RBAC permissions for it:

```
$ oc create sa faas-controller
$ oc policy add-role-to-user admin system:serviceaccount:openfaas:faas-controller --namespace=openfaas-fn
```

Note above that we are *not* using [rbac.yml](https://github.com/openfaas/faas-netes/blob/master/yaml/rbac.yml) to set up permissions.

Now it's time to deploy all OpenFaaS components:

```
$ cd yaml/
$ oc apply -f alertmanager_config.yml,alertmanager.yml,faasnetesd.yml,gateway.yml,nats.yml,prometheus_config.yml,prometheus.yml,queueworker.yml
$ oc expose service/gateway
```

You can check where the OpenFaaS UI is either via the OpenShift dashboard or by issuing the following command:

```
$ oc get routes
NAME      HOST/PORT                                PATH      SERVICES   PORT      TERMINATION   WILDCARD
gateway   gateway-openfaas.192.168.99.100.nip.io             gateway    8080                    None
```

In my case, the OpenFaaS UI is available via http://gateway-openfaas.192.168.99.100.nip.io so I'd go there and enjoy OpenFaaS!
