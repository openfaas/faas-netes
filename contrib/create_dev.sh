#!/usr/bin/env bash

# Creates an installation of OpenFaaS in kind for development

contrib/create_kind_cluster.sh
contrib/install_tiller.sh
contrib/deploy.sh
contrib/run_function.sh
