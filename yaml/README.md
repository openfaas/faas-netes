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
