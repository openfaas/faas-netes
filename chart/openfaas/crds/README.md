# CRDs

CRDs can be installed via `kubectl apply` before running the helm command, or through helm.

## 1. Use `kubectl` to manage CRDs

Use this option when Helm is not to be responsible for installing or updating CRDs.

```sh
export VERSION=14.2.43

cd /tmp/
mkdir -p chart

helm fetch openfaas/openfaas --version $VERSION --destination ./chart/openfaas --untar

kubectl apply -f ./chart/openfaas/openfaas/artifacts/crds.yaml
```

Then install the Helm chart. To update the CRDs, run the above with each update of the chart version.

## 2. Use `helm` to manage CRDs

When using helm to install CRDs, they are kept in two different locations: 

### 2.1 CRDs in the ./crds folder

These CRDs are installed once by Helm before installing the chart. This is required for the IAM CRDs because IAM CRs are created by the chart during the installation.

The initial installation is done by Helm, but subsequent updates must be applied manually:

```sh
export VERSION=14.2.43

cd /tmp/
mkdir -p chart

helm fetch openfaas/openfaas --version $VERSION --destination ./chart/openfaas --untar

kubectl apply -f ./chart/openfaas/openfaas/crds/
```

For a split-installation, skip installation by adding the `--skip-crds` flag to the `helm` command.

### 2.2 CRDs held in Helm templates

CRDs in the `./templates` folder are installed *and updated* by Helm. No further action is required when updating the chart version on the cluster.

### 3. Split installation of CRDs

If you want to install the CRDs separately from the main helm chart due to lacking a Cluster Admin role, change the following.

For the CRDs in the `./crds` folder, add `--skip-crds` to the helm command.

For the CRDs in the `./templates` folder, add `--set createCRDs=true` to the helm command, and add the `--show-only` flag:

```sh
export VERSION=14.2.43
cd /tmp/
mkdir -p chart

helm fetch openfaas/openfaas --version $VERSION --destination ./chart/openfaas --untar

helm template ./chart/openfaas/openfaas --set createCRDs=true \
    --show-only "templates/*-crd.yaml" > templates-crds.yaml
```

You can then apply `templates-crds.yaml` as required.

Later, when you install or upgrade the chart, add `--skip-crds=true --set createCRDs=false` to the helm command.

