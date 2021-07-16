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

{{/* Way to override KubeVersion. */}}
{{- define "openfaas.ingress.kubeVersion" -}}
  {{- default .Capabilities.KubeVersion.Version .Values.k8sVersionOverride -}}
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
