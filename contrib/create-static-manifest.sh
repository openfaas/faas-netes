#!/usr/bin/env bash

# Uses `helm template` to generate static yaml manifests for installation via `kubectl apply -f`.
#
# You can safely override the chart, the output directory, the values file used, the installation
# namespace. and the function namespace.
#
# This script is used to generate the default yaml installations included in the faas-netes repo.
#
# Usage:
#   $ ./create-static-manifest.sh
#   $ ./create-static-manifest.sh ./chart/openfaas ./yaml
#   $ ./create-static-manifest.sh ./chart/openfaas ./yaml value.yaml
#   $ ./create-static-manifest.sh ./chart/openfaas ./yaml value.yaml openfaas
#   $ ./create-static-manifest.sh ./chart/openfaas ./yaml value.yaml openfaas openfaas-fn

CHART_DIR=${1:-"./chart/openfaas"}
OUTPUT_DIR=${2:-"./yaml"}
VALUESNAME=${3:-"$CHART_DIR/values.yaml"}
NAMEPSPACE=${4:-"openfaas"}
FUNCTIONNAMESPACE=${5:-"openfaas-fn"}

TEMPLATE_FILE=$(find $CHART_DIR/templates/*.yaml -type f)
for filepath in $TEMPLATE_FILE; do
    filename=$(basename $filepath)
    outputname="${filename%%.*}.yml"
    # Use helm to generate the yaml and then use sed to remove the helm specific lables/annotations.
    helm template "$CHART_DIR" --name=openfaas --namespace="$NAMEPSPACE" --set "functionNamespace=$FUNCTIONNAMESPACE" -x templates/$filename --values="$VALUESNAME" \
        | sed -E '/(chart:)|(release:)|(heritage:)/d' \
        | sed -E '/# Source:/d' \
        | sed -E '/^$/d' \
        > $OUTPUT_DIR/$outputname \
        && echo "wrote $OUTPUT_DIR/$outputname"

    # Remove any  files that are exactly 4 characters, the are "empty" files, we expect that these
    # files will only contain "---\n"
    find $OUTPUT_DIR -size 4c -delete
done
