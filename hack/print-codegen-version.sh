#!/bin/bash

# This scripts exists primarily so that it can be used in the Makefile.
# It is needed because the `($shell ...)` command was having issues with the pipe.
# Extracting it to a script was the simplest solution.

grep 'k8s.io/code-generator' go.mod | awk '{print $2}'
