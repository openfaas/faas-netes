## Managing Secrets

Secrets come in two flavors, image-pull secrets and application secrets.

1.  Image-pull secrets that [Kubernetes will use to pull Docker images from a private registry](https://kubernetes.io/docs/tasks/configure-pod-container/pull-image-private-registry/).
2.  Application secrets are simply sensitive values that your function requires to function, e.g. a database password. `faas-netes` supports both types of secrets.

If you are familiar with the inner workings of Kubernetes, then image-pull secrets will be used as you would expect from the Kubernetes documentation. In contrast, application secrets will behave in the same way that secrets are handled in Docker Swarm, specifically, they will be mounted as files in the `/run/secrets` folder. This allows your functions to be agnostic of the deployment infrastructure. The relevant Kubernetes documentation for how this is implemented can be found [here](https://kubernetes.io/docs/concepts/configuration/secret/#using-secrets-as-files-from-a-pod).

_Note_: You can transparently use both types of secrets simultaneously for a function. `faas-netes` will gracefully inspect and handle secrets of different types.

## Image-Pull Secrets

### Deploy a function from a private Docker image

```
$ docker pull functions/alpine:latest
$ docker tag functions/alpine:latest $DOCKER_USERNAME/private-alpine:latest
$ docker push $DOCKER_USERNAME/private-alpine:latest
```

Now log into the [Hub](https://hub.docker.com) and make your image `private-alpine` private.

### Create your openfaas project

```sh
$ mkdir privatefuncs && cd privatefuncs
$ touch stack.yaml
```

In your favoriate editor, open `stack.yaml` and add

```yaml
provider:
  name: faas
  gateway: http://localhost:8080

functions:
  protectedapi:
    lang: Dockerfile
    skip_build: true
    image: username/private-alpine:latest
```

### Deploy the function

If you try to deploy using `faas-cli deploy` it will fail because Kubernetes can not pull the image.
You can verify this in the Kubernetes dashboard or via the CLI using the `kubectl describe` command.

To deploy the function, we need to create the [Image Pull Secret](https://kubernetes.io/docs/tasks/configure-pod-container/pull-image-private-registry/)

1.  Set the following environmental variables:

    * $DOCKER_USERNAME
    * $DOCKER_PASSWORD
    * $DOCKER_EMAIL

2.  Then run this command to create the needed secret

    ```sh
    $ kubectl create secret docker-registry dockerhub \
        --docker-username=$DOCKER_USERNAME \
        --docker-password=$DOCKER_PASSWORD \
        --docker-email=$DOCKER_EMAIL
    ```

3.  Update your stack file to include the secret:

    ```yaml
     provider:
       name: faas
       gateway: http://localhost:8080

     functions:
       protectedapi:
         lang: Dockerfile
         skip_build: true
         image: username/private-alpine:latest
         secrets:
          - dockerhub
    ```

Now you can again deploy the project using `faas-cli deploy`. Now when you inspect the Kubernetes pods, you will see that it can pull the docker image.

## Application Secrets

### Create a secret

What if that previous function also requires access to a database and that you have store the username and password credentials in `dbusername.txt` and `dbpass.txt`. Using the Kuberenetes documentation as a basis, we can create a secret in Kubernetes

```sh
kubectl create secret generic db-user-pass --from-file=./dbusername.txt --from-file=./dbpass.txt
secret "db-user-pass" created
```

### Update your yaml

```yaml
provider:
  name: faas
  gateway: http://localhost:8080

functions:
  protectedapi:
    lang: Dockerfile
    skip_build: true
    image: username/private-alpine:latest
    secrets:
    - dockerhub
    - db-user-pass
```

### Deploy the function

You can again use `faas-cli deploy` to deploy your function. Now, when you use `kubectl describe pod <pod-name>` you should see the `db-user-pass` referenced in the `Mounts` and `Volumes` sections of the output.
