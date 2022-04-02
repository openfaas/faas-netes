#!/bin/bash
export KUBEVAL_SCHEMA_LOCATION=${KUBEVAL_SCHEMA_LOCATION:-"https://raw.githubusercontent.com/yannh/kubernetes-json-schema/master/"}

ROOT=$(git rev-parse --show-toplevel)
KUBERNETES_VERSION=${KUBERNETES_VERSION:-'v1.21.1'}

status=0
msg=''

for version in $KUBERNETES_VERSION; do
    echo -e "\n#####################################"
    echo "Validating Kubernetes $version"
    echo "######################################"
    for chart in $ROOT/chart/*; do
        name=$(basename $chart)
        echo -e "\n\nValidating $name"
        helm template $chart -f $chart/values.yaml -n openfaas |
            kubeval -n openfaas --strict --ignore-missing-schemas --kubernetes-version "${version#v}" |
            grep -v 'WARN - Set to ignore missing schemas' |
            grep -v 'PASS'
        if [ $? != 0 ]; then
            status=1
            msg="$msg\n$name ($version)"
        fi
    done
done

if [ $status != 0 ]; then
    echo -e "Failures:\n$msg"
fi

exit $status
