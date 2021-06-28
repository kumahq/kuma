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

{{- define "kuma.controlPlane.globalZoneSync.serviceName" -}}
{{- $defaultSvcName := printf "%s-global-zone-sync" (include "kuma.name" .) -}}
{{ printf "%s" (default $defaultSvcName .Values.controlPlane.globalZoneSyncService.name) }}
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

{{- define "kuma.defaultEnv" -}}
env:
- name: KUMA_ENVIRONMENT
  value: "kubernetes"
- name: KUMA_STORE_TYPE
  value: "kubernetes"
- name: KUMA_STORE_KUBERNETES_SYSTEM_NAMESPACE
  value: {{ .Release.Namespace | quote }}
- name: KUMA_RUNTIME_KUBERNETES_CONTROL_PLANE_SERVICE_NAME
  value: {{ include "kuma.controlPlane.serviceName" . }}
- name: KUMA_GENERAL_TLS_CERT_FILE
  value: /var/run/secrets/kuma.io/tls-cert/tls.crt
- name: KUMA_GENERAL_TLS_KEY_FILE
  value: /var/run/secrets/kuma.io/tls-cert/tls.key
{{- if eq .Values.controlPlane.mode "zone" }}
- name: KUMA_MULTIZONE_ZONE_GLOBAL_ADDRESS
  value: {{ .Values.controlPlane.kdsGlobalAddress }}
{{- end }}
- name: KUMA_DP_SERVER_HDS_ENABLED
  value: "false"
- name: KUMA_API_SERVER_READ_ONLY
  value: "true"
- name: KUMA_RUNTIME_KUBERNETES_ADMISSION_SERVER_PORT
  value: "5443"
- name: KUMA_RUNTIME_KUBERNETES_ADMISSION_SERVER_CERT_DIR
  value: /var/run/secrets/kuma.io/tls-cert
- name: KUMA_RUNTIME_KUBERNETES_INJECTOR_CNI_ENABLED
  value: {{ .Values.cni.enabled | quote }}
- name: KUMA_RUNTIME_KUBERNETES_INJECTOR_SIDECAR_CONTAINER_IMAGE
  value: {{ include "kuma.formatImage" (dict "image" .Values.dataPlane.image "root" $) | quote }}
- name: KUMA_INJECTOR_INIT_CONTAINER_IMAGE
  value: {{ include "kuma.formatImage" (dict "image" .Values.dataPlane.initImage "root" $) | quote }}
- name: KUMA_RUNTIME_KUBERNETES_INJECTOR_CA_CERT_FILE
  value: /var/run/secrets/kuma.io/tls-cert/ca.crt
- name: KUMA_DEFAULTS_SKIP_MESH_CREATION
  value: {{ .Values.controlPlane.defaults.skipMeshCreation | quote }}
- name: KUMA_MODE
  value: {{ .Values.controlPlane.mode | quote }}
{{- if .Values.controlPlane.zone }}
- name: KUMA_MULTIZONE_ZONE_NAME
  value: {{ .Values.controlPlane.zone | quote }}
{{- end }}
{{- if .Values.controlPlane.tls.apiServer.secretName }}
- name: KUMA_API_SERVER_HTTPS_TLS_CERT_FILE
  value: /var/run/secrets/kuma.io/api-server-tls-cert/tls.crt
- name: KUMA_API_SERVER_HTTPS_TLS_KEY_FILE
  value: /var/run/secrets/kuma.io/api-server-tls-cert/tls.key
{{- end }}
{{- if .Values.controlPlane.tls.apiServer.clientCertsSecretName }}
- name: KUMA_API_SERVER_AUTH_CLIENT_CERTS_DIR
  value: /var/run/secrets/kuma.io/api-server-client-certs/
{{- end }}
{{- if .Values.controlPlane.tls.kdsGlobalServer.secretName }}
- name: KUMA_MULTIZONE_GLOBAL_KDS_TLS_CERT_FILE
  value: /var/run/secrets/kuma.io/kds-server-tls-cert/tls.crt
- name: KUMA_MULTIZONE_GLOBAL_KDS_TLS_KEY_FILE
  value: /var/run/secrets/kuma.io/kds-server-tls-cert/tls.key
{{- end }}
{{- if .Values.controlPlane.tls.kdsZoneClient.secretName }}
- name: KUMA_MULTIZONE_ZONE_KDS_ROOT_CA_FILE
  value: /var/run/secrets/kuma.io/kds-client-tls-cert/ca.crt
{{- end }}
{{- end }}

{{/*
params: { image: { registry?, repository, tag? }, root: $ }
returns: formatted image string
*/}}
{{- define "kubectl.formatImage" -}}
{{- $img := .image }}
{{- $tag := .tag }}
{{- $root := .root }}
{{- $registry := ($img.registry | default $root.Values.kubectl.image.registry) -}}
{{- $repo := ($img.repository | default $root.Values.kubectl.image.repository) -}}
{{- $imageTag := ($tag | default $root.Values.kubectl.image.tag) -}}
{{- printf "%s/%s:%s" $registry $repo $imageTag -}}
{{- end -}}
