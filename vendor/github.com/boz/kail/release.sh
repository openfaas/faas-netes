#!/bin/bash

set -e
set -o pipefail

if [ "$TRAVIS_REPO_SLUG"  != "boz/kail" -o \
     "$TRAVIS_EVENT_TYPE" != "push"     -o \
     "$TRAVIS_GO_VERSION" != "1.10"      ]; then
  exit 0
fi

if [ "$TRAVIS_BRANCH"    != "master" -a \
     "${TRAVIS_TAG:0:1}" != "v"         ]; then
  exit 0
fi

docker login -u "$DOCKERHUB_USERNAME" -p "$DOCKERHUB_PASSWORD"

if [ "$TRAVIS_BRANCH" == "master" ]; then
  make image-push
  exit 0
fi

DOCKER_TAG="$TRAVIS_TAG" make image-push

curl -sL https://git.io/goreleaser |     \
  GITHUB_TOKEN="$GITHUB_REPO_TOKEN" bash
