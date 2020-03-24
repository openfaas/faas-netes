echo Installing with helm 👑

helm repo add openfaas https://openfaas.github.io/faas-netes/

kubectl apply -f \
   https://raw.githubusercontent.com/openfaas/faas-netes/master/namespaces.yml

# generate a random password
PASSWORD=$(head -c 12 /dev/urandom | shasum| cut -d' ' -f1)

kubectl -n openfaas create secret generic basic-auth \
--from-literal=basic-auth-user=admin \
--from-literal=basic-auth-password="$PASSWORD"

echo "Installing chart 🍻"
helm upgrade \
    --install \
    openfaas \
    openfaas/openfaas \
    --namespace openfaas  \
    --set basic_auth=true \
    --set functionNamespace=openfaas-fn \
    --set serviceType=LoadBalancer \
    --wait
