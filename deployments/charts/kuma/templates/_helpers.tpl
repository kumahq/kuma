{{/* vim: set filetype=mustache: */}}
{{/*
Expand the name of the chart.
*/}}
{{- define "kuma.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
This is the Kuma version the chart is intended to be used with.
*/}}
{{- define "kuma.appVersion" -}}
{{- .Chart.AppVersion -}}
{{- end }}

{{/*
This is only used in the `kuma.formatImage` function below.
*/}}
{{- define "kuma.defaultRegistry" -}}
docker.io/kumahq
{{- end }}

{{- define "kuma.product" -}}
Kuma
{{- end }}

{{- define "kuma.tagPrefix" -}}
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

{{- define "kuma.egress.serviceName" -}}
{{- $defaultSvcName := printf "%s-egress" (include "kuma.name" .) -}}
{{ printf "%s" (default $defaultSvcName .Values.egress.service.name) }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "kuma.labels" -}}
helm.sh/chart: {{ include "kuma.chart" . }}
{{ include "kuma.selectorLabels" . }}
{{- if (include "kuma.appVersion" .) }}
app.kubernetes.io/version: {{ (include "kuma.appVersion" .) | quote }}
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
control plane labels
*/}}
{{- define "kuma.cpLabels" -}}
app: {{ include "kuma.name" . }}-control-plane
{{- range $key, $value := $.Values.controlPlane.extraLabels }}
{{ $key | quote }}: {{ $value | quote }}
{{- end }}
{{ include "kuma.labels" . }}
{{- end }}

{{/*
ingress labels
*/}}
{{- define "kuma.ingressLabels" -}}
app: {{ include "kuma.name" . }}-ingress
{{- range $key, $value := .Values.ingress.extraLabels }}
{{ $key | quote }}: {{ $value | quote }}
{{- end }}
{{ include "kuma.labels" . }}
{{- end }}

{{/*
egress labels
*/}}
{{- define "kuma.egressLabels" -}}
app: {{ include "kuma.name" . }}-egress
{{ range $key, $value := .Values.egress.extraLabels }}
{{ $key | quote }}: {{ $value | quote }}
{{ end }}
{{- include "kuma.labels" . }}
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
{{- $product := (include "kuma.product" .) }}
{{- $tagPrefix := (include "kuma.tagPrefix" .) }}
{{- $expectedVersion := (include "kuma.appVersion" $root) }}
{{- if
  and
    $root.Values.global.image.tag
    (ne $root.Values.global.image.tag (include "kuma.appVersion" $root))
    (eq $root.Values.global.image.registry (include "kuma.defaultRegistry" .))
-}}
{{- fail (
  printf "This chart only supports %s version %q but %sglobal.image.tag is set to %q. Set %sglobal.image.tag to %q or skip this check by setting %s*.image.tag for each individual component."
  $product $expectedVersion $tagPrefix $root.Values.global.image.tag $tagPrefix $expectedVersion $tagPrefix
) -}}
{{- end -}}
{{- $defaultTag := ($root.Values.global.image.tag | default (include "kuma.appVersion" $root)) -}}
{{- $tag := ($img.tag | default $defaultTag) -}}
{{- printf "%s/%s:%s" $registry $repo $tag -}}
{{- end -}}

{{- define "kuma.defaultEnv" -}}
{{ if not (or (eq .Values.controlPlane.mode "zone") (eq .Values.controlPlane.mode "global") (eq .Values.controlPlane.mode "standalone")) }}
  {{ $msg := printf "controlPlane.mode invalid got:'%s' supported values: global,zone,standalone" .Values.controlPlane.mode }}
  {{ fail $msg }}
{{ end }}
{{ if eq .Values.controlPlane.mode "zone" }}
  {{ if empty .Values.controlPlane.zone }}
    {{ fail "Can't have controlPlane.zone to be empty when controlPlane.mode=='zone'" }}
  {{ end }}
  {{ if empty .Values.controlPlane.kdsGlobalAddress }}
    {{ fail "controlPlane.kdsGlobalAddress can't be empty when controlPlane.mode=='zone', needs to be the global control-plane address" }}
  {{ else }}
    {{ $url := urlParse .Values.controlPlane.kdsGlobalAddress }}
    {{ if not (eq $url.scheme "grpcs") }}
      {{ $msg := printf "controlPlane.kdsGlobalAddress must be a url with scheme grpcs:// got:'%s'" .Values.controlPlane.kdsGlobalAddress }}
      {{ fail $msg }}
    {{ end }}
  {{ end }}
{{ else }}
  {{ if not (empty .Values.controlPlane.zone) }}
    {{ fail "Can't specify a controlPlane.zone when controlPlane.mode!='zone'" }}
  {{ end }}
  {{ if not (empty .Values.controlPlane.kdsGlobalAddress) }}
    {{ fail "Can't specify a controlPlane.kdsGlobalAddress when controlPlane.mode!='zone'" }}
  {{ end }}
{{ end }}
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
{{- if and (eq .Values.controlPlane.mode "global") (or .Values.controlPlane.tls.kdsGlobalServer.secretName .Values.controlPlane.tls.kdsGlobalServer.create) }}
- name: KUMA_MULTIZONE_GLOBAL_KDS_TLS_CERT_FILE
  value: /var/run/secrets/kuma.io/kds-server-tls-cert/tls.crt
- name: KUMA_MULTIZONE_GLOBAL_KDS_TLS_KEY_FILE
  value: /var/run/secrets/kuma.io/kds-server-tls-cert/tls.key
{{- end }}
{{- if and (eq .Values.controlPlane.mode "zone") (or .Values.controlPlane.tls.kdsZoneClient.secretName .Values.controlPlane.tls.kdsZoneClient.create) }}
- name: KUMA_MULTIZONE_ZONE_KDS_ROOT_CA_FILE
  value: /var/run/secrets/kuma.io/kds-client-tls-cert/ca.crt
{{- end }}
- name: KUMA_API_SERVER_AUTHN_LOCALHOST_IS_ADMIN
  value: "false"
- name: KUMA_RUNTIME_KUBERNETES_SERVICE_ACCOUNT_NAME
  value: "system:serviceaccount:{{ .Release.Namespace }}:{{ include "kuma.name" . }}-control-plane"
{{- if .Values.experimental.gatewayAPI }}
- name: KUMA_EXPERIMENTAL_GATEWAY_API
  value: "true"
{{- end }}
{{- if .Values.experimental.cni }}
- name: KUMA_RUNTIME_KUBERNETES_NODE_TAINT_CONTROLLER_ENABLED
  value: "true"
- name: KUMA_RUNTIME_KUBERNETES_NODE_TAINT_CONTROLLER_CNI_APP
  value: "{{ include "kuma.name" . }}-cni"
{{- end }}
{{- if .Values.experimental.transparentProxy }}
- name: KUMA_RUNTIME_KUBERNETES_INJECTOR_TRANSPARENT_PROXY_V2
  value: "true"
{{- end }}
{{- if .Values.experimental.ebpf.enabled }}
- name: KUMA_RUNTIME_KUBERNETES_INJECTOR_EBPF_ENABLED
  value: "true"
- name: KUMA_RUNTIME_KUBERNETES_INJECTOR_EBPF_INSTANCE_IP_ENV_VAR_NAME
  value: {{ .Values.experimental.ebpf.instanceIPEnvVarName }}
- name: KUMA_RUNTIME_KUBERNETES_INJECTOR_EBPF_BPFFS_PATH
  value: {{ .Values.experimental.ebpf.bpffsPath }}
- name: KUMA_RUNTIME_KUBERNETES_INJECTOR_EBPF_PROGRAMS_SOURCE_PATH
  value: {{ .Values.experimental.ebpf.programsSourcePath }}
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
