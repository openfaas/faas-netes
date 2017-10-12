## Managing Secrets

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

