{{/* vim: set filetype=mustache: */}}
{{/*
Expand the name of the chart.
*/}}
{{- define "openfaas.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
*/}}
{{- define "openfaas.fullname" -}}
{{- $name := default .Chart.Name .Values.nameOverride -}}
{{- if contains $name .Release.Name -}}
{{- .Release.Name | trunc 63 | trimSuffix "-" -}}
{{- else -}}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" -}}
{{- end -}}
{{- end -}}

{{/* Way to override KubeVersion, with EKS suffix sanitization */}}
{{- define "openfaas.ingress.kubeVersion" -}}
  {{- $version := default .Capabilities.KubeVersion.Version .Values.k8sVersionOverride -}}
  {{- regexReplaceAll "([0-9]+\\.[0-9]+\\.[0-9]+).*" $version "$1" -}}
{{- end -}}

{{/* Determine Ingress API Version */}}
{{- define "openfaas.ingress.apiVersion" -}}
  {{- if and (.Capabilities.APIVersions.Has "networking.k8s.io/v1") (semverCompare ">= 1.19.x" (include "openfaas.ingress.kubeVersion" .)) -}}
      {{- print "networking.k8s.io/v1" -}}
  {{- else if .Capabilities.APIVersions.Has "networking.k8s.io/v1beta1" -}}
    {{- print "networking.k8s.io/v1beta1" -}}
  {{- else -}}
    {{- print "extensions/v1beta1" -}}
  {{- end -}}
{{- end -}}

{{/* Check Ingress stability */}}
{{- define "openfaas.ingress.isStable" -}}
  {{- eq (include "openfaas.ingress.apiVersion" .) "networking.k8s.io/v1" -}}
{{- end -}}

{{/* Check Ingress pathType support > started with Kubernetes 1.18 */}}
{{- define "openfaas.ingress.supportsPathType" -}}
  {{- or (eq (include "openfaas.ingress.isStable" .) "true") (and (eq (include "openfaas.ingress.apiVersion" .) "networking.k8s.io/v1beta1") (semverCompare ">= 1.18.x" (include "openfaas.ingress.kubeVersion" .))) -}}
{{- end -}}

{{/*
Image helper that replaces registry with custom prefix if specified.
Usage: {{ include "openfaas.image" (dict "image" .Values.someImage "registryPrefix" .Values.registryPrefix) }}
*/}}
{{- define "openfaas.image" -}}
{{- $image := .image -}}
{{- $registryPrefix := .registryPrefix -}}
{{- if $registryPrefix -}}
  {{- if hasPrefix "docker.io/" $image -}}
    {{- printf "%s/%s" $registryPrefix (trimPrefix "docker.io/" $image) -}}
  {{- else if hasPrefix "ghcr.io/" $image -}}
    {{- printf "%s/%s" $registryPrefix (trimPrefix "ghcr.io/" $image) -}}
  {{- else if hasPrefix "quay.io/" $image -}}
    {{- printf "%s/%s" $registryPrefix (trimPrefix "quay.io/" $image) -}}
  {{- else if hasPrefix "registry.k8s.io/" $image -}}
    {{- printf "%s/%s" $registryPrefix (trimPrefix "registry.k8s.io/" $image) -}}
  {{- else if contains "/" $image -}}
    {{- /* Image has a registry, replace the first part */ -}}
    {{- $parts := splitList "/" $image -}}
    {{- if gt (len $parts) 2 -}}
      {{- /* More than 2 parts, replace first part */ -}}
      {{- printf "%s/%s" $registryPrefix (join "/" (rest $parts)) -}}
    {{- else -}}
      {{- /* Exactly 2 parts, keep the organization/repository structure */ -}}
      {{- printf "%s/%s" $registryPrefix $image -}}
    {{- end -}}
  {{- else -}}
    {{- /* No registry prefix, add our custom one */ -}}
    {{- printf "%s/%s" $registryPrefix $image -}}
  {{- end -}}
{{- else -}}
  {{- $image -}}
{{- end -}}
{{- end -}}
