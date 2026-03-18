#!/bin/bash

## Example for rendering Helm chart for openfaas to a static
## set of objects with predictable/stable names for easier
## diffs and maintenance for teams who need to go through a manual
## YAML approval for each change/object vs reviewing the chart directly

OUTPUT="yaml/openfaas/"
rm -rf "$OUTPUT"

helm template chart/openfaas/ -f ./chart/openfaas/values-pro.yaml | \
  yq -N -s '"__split__" + $index' - && \
for f in __split__*.yml; do
  kind=$(yq '.kind' "$f")
  name=$(yq '.metadata.name' "$f")

  if [ "$kind" = "null" ] || [ -z "$kind" ]; then
    rm -f "$f"
    continue
  fi

  lower_kind=$(echo "$kind" | tr '[:upper:]' '[:lower:]')

  prefix=""
  if [ "$kind" = "CustomResourceDefinition" ]; then
    prefix="00_"
  fi

  dir="${OUTPUT}/${prefix}${lower_kind}"
  mkdir -p "$dir"
  mv "$f" "${dir}/${name}.yaml"
done
