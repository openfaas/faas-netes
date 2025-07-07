#!/bin/bash
ROOT=$(git rev-parse --show-toplevel)

status=0
msg=''

for chart in $ROOT/chart/*; do
  name=$(basename $chart)
  echo -e "\n\nLinting $name"
  helm lint "$chart"
  if [ $? != 0 ]; then
    status=1
    msg="$msg\n$name"
  fi
done

if [ $status != 0 ]; then
  echo -e "Helm lint failures:\n$msg"
  exit $status
fi


# Run kubeconform on all charts as an extra measure

for chart in $ROOT/chart/*; do
  name=$(basename $chart)
  echo -e "\n\nKubeconform $name"
  # Use ignore-missing-schemas to avoid errors with missing CRD schemas for the openfaas chart only

  echo "Using kubeconform on $name - $chart"

  if [ "$name" == "openfaas" ]; then
    helm template chart/$name -f ./chart/openfaas/values-pro.yaml | kubeconform --ignore-missing-schemas -summary --verbose --strict
  else
    helm template chart/$name | kubeconform -summary --verbose --strict
  fi
  
  if [ $? != 0 ]; then
    status=1
    msg="$msg\n$name"
  fi
done

if [ $status != 0 ]; then
  echo -e "Kubeconform failures:\n$msg"
  exit $status
fi

echo -e "\n\nAll charts passed linting and kubeconform validation"
exit 0

