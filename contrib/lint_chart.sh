#!/bin/bash

for chart in `ls ./chart`; do
  helm lint ./chart/${chart}
  if [ $? != 0 ]; then
    exit 1
  fi
done

