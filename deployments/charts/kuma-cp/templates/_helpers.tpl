{{/* vim: set filetype=mustache: */}}
{{/*
Expand the name of the chart.
*/}}
{{- define "kuma-cp.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "kuma-cp.fullname" -}}
{{- if .Values.fullnameOverride }}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- $name := default .Chart.Name .Values.nameOverride }}
{{- if contains $name .Release.Name }}
{{- .Release.Name | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" }}
{{- end }}
{{- end }}
{{- end }}

{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "kuma-cp.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{- define "kuma-cp.serviceName" -}}
{{- $defaultSvcName := printf "%s-control-plane" (include "kuma-cp.name" .) -}}
{{ printf "%s" (default $defaultSvcName .Values.controlPlane.service.name) }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "kuma-cp.labels" -}}
helm.sh/chart: {{ include "kuma-cp.chart" . }}
{{ include "kuma-cp.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "kuma-cp.selectorLabels" -}}
app.kubernetes.io/name: {{ include "kuma-cp.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
CNI labels
*/}}
{{- define "kuma-cp.cniLabels" -}}
app: {{ include "kuma-cp.name" . }}-cni
{{ include "kuma-cp.labels" . }}
{{- end }}

{{/*
CNI selector labels
*/}}
{{- define "kuma-cp.cniSelectorLabels" -}}
app: {{ include "kuma-cp.name" . }}-cni
{{ include "kuma-cp.selectorLabels" . }}
{{- end }}

{{/*
params: { image: { registry?, repository, tag? }, root: $ }
returns: formatted image string
*/}}
{{- define "kuma-cp.formatImage" -}}
{{- $img := .image }}
{{- $root := .root }}
{{- $registry := ($img.registry | default $root.Values.global.image.registry) -}}
{{- $repo := ($img.repository | required "Must specify image repository") -}}
{{- $defaultTag := ($root.Values.global.image.tag | default $root.Chart.AppVersion) -}}
{{- $tag := ($img.tag | default $defaultTag) -}}
{{- printf "%s/%s:%s" $registry $repo $tag -}}
{{- end -}}
