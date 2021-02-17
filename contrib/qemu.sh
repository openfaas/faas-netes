#!/bin/bash

# Support multi-arch builds with buildx on a Linux host.
docker run --rm --privileged multiarch/qemu-user-static --reset -p yes
