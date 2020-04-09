#!/usr/bin/env bash

# Creates an installation of OpenFaaS in kind for development

contrib/create_cluster.sh
contrib/deploy.sh
contrib/run_function.sh
