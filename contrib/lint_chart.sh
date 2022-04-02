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
  echo -e "Failures:\n$msg"
fi

exit $status
