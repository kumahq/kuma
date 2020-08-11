{{/* vim: set filetype=mustache: */}}
{{/*
Expand the name of the chart.
*/}}
{{- define "kuma.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "kuma.fullname" -}}
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
{{- define "kuma.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{- define "kuma.controlPlane.serviceName" -}}
{{- $defaultSvcName := printf "%s-control-plane" (include "kuma.name" .) -}}
{{ printf "%s" (default $defaultSvcName .Values.controlPlane.service.name) }}
{{- end }}

{{- define "kuma.controlPlane.globalRemoteSync.serviceName" -}}
{{- $defaultSvcName := printf "%s-global-remote-sync" (include "kuma.name" .) -}}
{{ printf "%s" (default $defaultSvcName .Values.controlPlane.globalRemoteSyncService.name) }}
{{- end }}

{{- define "kuma.ingress.serviceName" -}}
{{- $defaultSvcName := printf "%s-ingress" (include "kuma.name" .) -}}
{{ printf "%s" (default $defaultSvcName .Values.ingress.service.name) }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "kuma.labels" -}}
helm.sh/chart: {{ include "kuma.chart" . }}
{{ include "kuma.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "kuma.selectorLabels" -}}
app.kubernetes.io/name: {{ include "kuma.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
CNI labels
*/}}
{{- define "kuma.cniLabels" -}}
app: {{ include "kuma.name" . }}-cni
{{ include "kuma.labels" . }}
{{- end }}

{{/*
CNI selector labels
*/}}
{{- define "kuma.cniSelectorLabels" -}}
app: {{ include "kuma.name" . }}-cni
{{ include "kuma.selectorLabels" . }}
{{- end }}

{{/*
params: { image: { registry?, repository, tag? }, root: $ }
returns: formatted image string
*/}}
{{- define "kuma.formatImage" -}}
{{- $img := .image }}
{{- $root := .root }}
{{- $registry := ($img.registry | default $root.Values.global.image.registry) -}}
{{- $repo := ($img.repository | required "Must specify image repository") -}}
{{- $defaultTag := ($root.Values.global.image.tag | default $root.Chart.AppVersion) -}}
{{- $tag := ($img.tag | default $defaultTag) -}}
{{- printf "%s/%s:%s" $registry $repo $tag -}}
{{- end -}}
