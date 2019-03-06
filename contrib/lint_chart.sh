#!/bin/bash

for chart in `ls ./chart`; do
  helm lint ./chart/${chart}
  if [ $? != 0 ]; then
    travis_terminate 1
  fi
done

