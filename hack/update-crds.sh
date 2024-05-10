#!/bin/bash

GOPATH=$(go env GOPATH)
controllergen="$GOPATH/bin/controller-gen"
PKG=sigs.k8s.io/controller-tools/cmd/controller-gen@v0.14.0

if [ ! -e "$controllergen" ]; then
  echo "Getting $PKG"
  go install $PKG
fi

"$controllergen" \
  crd \
  schemapatch:manifests=./artifacts/crds \
  paths=./pkg/apis/... \
  output:dir=./artifacts/crds


# Prevent CRDs from being removed via helm uninstall

# Only annotate CRDs which are held in templates, but not the ones in the crds/ folder
ANNOTATE_CRDS=("openfaas.com_profiles.yaml" "openfaas.com_functioningresses.yaml" "openfaas.com_functions.yaml")

for f in ./artifacts/crds/*.yaml; do \

  annotate=false
  for i in "${ANNOTATE_CRDS[@]}"; do
    if [ "$i" == "$(basename $f)" ]; then
      annotate=true
      break
    fi
  done

  echo "Updating: $f"
  tmp_file=$(mktemp)

  if [ "$annotate" == true ]; then
    kubectl annotate --overwrite -f $f --local=true -o yaml helm.sh/resource-policy=keep > tmp_file
    mv tmp_file $f
  fi

  kubectl label --overwrite -f $f --local=true -o yaml app.kubernetes.io/name=openfaas > tmp_file
  mv tmp_file $f

  rm -f tmp_file
done

echo "# Combined OpenFaaS CRDs " > ./artifacts/crds/crds.yaml

for f in ./artifacts/crds/*.yaml; do \
  if [ $f != "./artifacts/crds/crds.yaml" ]; then
    echo "---" >> ./artifacts/crds/crds.yaml
    cat $f >> ./artifacts/crds/crds.yaml
    echo "" >> ./artifacts/crds/crds.yaml

  fi
done

echo "Wrote combined CRDs to ./chart/openfaas/artifacts/crds.yaml"

cp ./artifacts/crds/crds.yaml ./chart/openfaas/artifacts/

echo "Copied IAM CRDs to ./chart/openfaas/crds/"
cp ./artifacts/crds/iam.*.yaml ./chart/openfaas/crds/