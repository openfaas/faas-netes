#!/usr/bin/env bash

# Usage: ./create-static-manifest.sh ./chart/openfaas ./yaml

CHART_DIR=${1:-"./chart/openfaas"}
OUTPUT_DIR=${2:-"./yaml"}
VALUESNAME=${3:-"values.yaml"}
NAMEPSPACE=${4:-"openfaas"}

TEMPLATE_FILE=$(find $CHART_DIR/templates/*.yaml -type f)
for filepath in $TEMPLATE_FILE; do
    filename=$(basename $filepath)
    outputname="${filename%%.*}.yml"
    helm template "$CHART_DIR" --name=faas --namespace="$NAMEPSPACE" -x templates/$filename --values="$CHART_DIR/$VALUESNAME" \
        | sed -E '/(chart:)|(release:)|(heritage:)/d' \
        | sed -E '/# Source:/d' \
        | sed -E '/^$/d' \
        > $OUTPUT_DIR/$outputname \
        && echo "wrote $OUTPUT_DIR/$outputname"

    find $OUTPUT_DIR -size 4c -delete
done
