## Managing Secrets

Secrets come in two flavors, image-pull secrets and application secrets.  
    1. Image-pull secrets that [Kubernetes will use to pull Docker images from a private registry](https://kubernetes.io/docs/tasks/configure-pod-container/pull-image-private-registry/).  
    2. Application secrets are simply sensitive values that your function requires to function, e.g. a database password. `faas-netes` supports both types of secrets.  
If you are familiar with the inner workings of Kubernetes, then image-pull secrets will be used as you would expect from the Kubernetes documentation. In contrast, application secrets will behave in the same way that secrets are handled in Docker Swarm, specifically, they will be mounted as files in the `/run/secrets` folder.  This allows your functions to be agnostic of the deployment infrastructure. The relevant Kubernetes documentation for how this is implemented can be found [here](https://kubernetes.io/docs/concepts/configuration/secret/#using-secrets-as-files-from-a-pod).

_Note_: You can transparently use both types of secrets simultaneously for a function. `faas-netes` will gracefully inspect and handle secrets of different types.

## Image-Pull Secrets

### Create an image-pull secret:

Set the following environmental variables:

* $DOCKER_USERNAME
* $DOCKER_PASSWORD
* $DOCKER_EMAIL

```
$ kubectl create secret docker-registry dockerhub \
    --docker-username=$DOCKER_USERNAME \
    --docker-password=$DOCKER_PASSWORD \
    --docker-email=$DOCKER_EMAIL
```

### Deploy a function from a private Docker image

```
$ docker pull functions/alpine:latest
$ docker tag functions/alpine:latest $DOCKER_USERNAME/private-alpine:latest
$ docker push $DOCKER_USERNAME/private-alpine:latest
```

Now log into the Hub and make your image `private-alpine` private.

### Attempt to deploy the image

You should see that the pod is not scheduled.

```
curl -X POST -d \
'{"image": "alexellis2/private-alpine:latest", "envProcess": "sha512sum", "service": "shametoo", "secrets": []}' $(minikube ip):31111/system/functions
```

### Now invoke the faas-netes back-end directly

Now update the function to use the new secret:


```
curl -X POST -d \
'{"image": "alexellis2/private-alpine:latest", "envProcess": "sha512sum", "service": "shametoo", "secrets": ["dockerhub"]}' $(minikube ip):31111/system/functions
```

You will now see that your function from a private registry has been deployed.

## Application Secrets

### Create a secret
Consider a function that requires access to a database and that you have store the username and password credentials in `dbusername.txt` and `dbpass.txt`. Using the Kuberenetes documentation as a basis, we can create a secret in Kubernetes

```sh
kubectl create secret generic db-user-pass --from-file=./dbusername.txt --from-file=./dbpass.txt
secret "db-user-pass" created
```

### Deploy the your function

Just as in the example for image-pull secrets, we simply add the name of the secret to the deploy request.

```
curl -X POST -d \
'{"image": "yourexample/db-fnc-alpine:latest", "envProcess": "./secretFunc", "service": "secret-func", "secrets": ['db-user-pass']}' $(minikube ip):31111/system/functions
```


