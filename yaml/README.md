# YAML reference

These instructions are for OpenFaaS when you don't want to use `helm`.

## 1.0 Enable basic auth on gateway

To enable built-in basic auth: create secrets, configure and mount the secrets to the gateway and queue worker.

### 1.1 Create secrets:

```
export GW_PASS=$(head -c 16 /dev/random |shasum | cut -d ' ' -f1)

echo "Password: $GW_PASS"

kubectl create secret generic gateway-basic-auth -n openfaas \
    --from-literal=basic-auth-user=admin \
    --from-literal=basic-auth-password=$GW_PASS

kubectl create secret generic basic-auth-user -n openfaas-fn  \
    --from-literal=basic-auth-user=admin

kubectl create secret generic basic-auth-password -n openfaas-fn \
 --from-literal=basic-auth-password=$GW_PASS
```

Record the line from "Password: c66ec59d880e66fb02c043ef910eefba020f2a3a" for later use with `faas-cli login`

### 1.2 Set env-vars on gateway and queue worker

```
        - name: basic_auth
          value: "true"
```

### 1.3 Add volume mount to gateway and queue worker

```
        volumeMounts:
        - name: gateway-basic-auth
          readOnly: true
          mountPath: "/etc/openfaas"
```

### 1.4 Add secrets as volumes within gateway and queue worker
```
      volumes:
      - name: gateway-basic-auth
        secret:
          secretName: gateway-basic-auth
```

Apply the changes to the gateway and queue worker(deployment only):

```
$ kubectl apply -f ./yaml/gateway-dep.yml
$ kubectl apply -f ./yaml/queueworker-dep.yml
```

### 1.5 Login in

Now attempt to login with:

```
echo -n $GW_PASS | faas-cli login --username=admin --password-stdin
```

## 2.0 Enable Non-repudiation for asynchronous function callbacks

To enable built-in Non-repudiation for asynchronous function callbacks: create secrets, 
configure and mount the secrets to the gateway and queue worker.

### 2.1 Create secrets:

```
rm signing.key > /dev/null 2>&1 || true && rm signing.key.pub > /dev/null 2>&1 || true
ssh-keygen -t rsa -b 2048 -N "" -m PEM -f signing.key > /dev/null 2>&1
openssl rsa -in ./signing.key -pubout -outform PEM -out signing.key.pub > /dev/null 2>&1

kubectl create secret generic http-signing-private-key -n openfaas \
    --from-file=http-signing-private-key=./signing.key

kubectl create secret generic http-signing-public-key -n openfaas \
    --from-file=http-signing-public-key=./signing.key.pub
 
rm signing.key || true && rm signing.key.pub || true
```

### 2.2 Add volume mount to gateway

```
        volumeMounts:
        - name: http-signing-public-key
          readOnly: true
          mountPath: "/etc/openfaas"
```

### 2.3 Add secrets as volumes to gateway

```
      volumes:
      - name: http-signing-public-key
        secret:
          secretName: http-signing-public-key
```


### 2.4 Add volume mount to queue worker

```
        volumeMounts:
        - name: http-signing-private-key
          readOnly: true
          mountPath: "/etc/openfaas"
```

### 2.5 Add secrets as volumes to queue worker

```
      volumes:
      - name: http-signing-private-key
        secret:
          secretName: http-signing-private-key
```

Apply the changes to the gateway and queue worker(deployment only):

```
$ kubectl apply -f ./yaml/gateway-dep.yml
$ kubectl apply -f ./yaml/queueworker-dep.yml
```