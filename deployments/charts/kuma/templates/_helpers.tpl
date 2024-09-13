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
control plane deployment annotations
*/}}
{{- define "kuma.cpDeploymentAnnotations" -}}
{{- range $key, $value := $.Values.controlPlane.deploymentAnnotations }}
{{ $key | quote }}: {{ $value | quote }}
{{- end }}
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
params: { dns: { policy?, config: {nameservers?, searches?}} }
returns: formatted dnsConfig
*/}}
{{- define "kuma.dnsConfig" -}}
{{- $dns := .dns }}
{{- if $dns.policy }}
dnsPolicy: {{ $dns.policy }}
{{- end }}
{{- if or (gt (len $dns.config.nameservers) 0) (gt (len $dns.config.searches) 0) }}
dnsConfig:
  {{- if gt (len $dns.config.nameservers) 0 }}
  nameservers:
    {{- range $nameserver := $dns.config.nameservers }}
    - {{ $nameserver }}
    {{- end }}
  {{- end }}
  {{- if gt (len $dns.config.searches) 0 }}
  searches:
    {{- range $search := $dns.config.searches }}
    - {{ $search }}
    {{- end }}
  {{- end }}
{{- end }}
{{- end -}}

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

{{- define "kuma.parentEnv" -}}
{{- end -}}

{{- define "kuma.parentSecrets" -}}
{{- end -}}

{{- define "kuma.pluginPoliciesEnabled" -}}
{{- $list := list -}}
{{- range $k, $v := .Values.plugins.policies -}}
{{- if $v -}}
{{- $list = append $list (printf "%s" $k) -}}
{{- end -}}
{{- end -}}
{{ join "," $list }}
{{- end -}}

{{- define "kuma.transparentProxyConfigMapName" -}}
{{- if .Values.transparentProxy.configMap.name }}
{{- .Values.transparentProxy.configMap.name | trunc 253 | trimSuffix "-" }}
{{- else }}
{{- printf "%s-transparent-proxy-config" .Chart.Name }}
{{- end }}
{{- end }}

{{- define "kuma.defaultEnv" -}}
env:
{{ include "kuma.parentEnv" . }}
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
  value: {{ .Values.controlPlane.admissionServerPort | default "5443" | quote }}
- name: KUMA_RUNTIME_KUBERNETES_ADMISSION_SERVER_CERT_DIR
  value: /var/run/secrets/kuma.io/tls-cert
- name: KUMA_RUNTIME_KUBERNETES_INJECTOR_CNI_ENABLED
  value: {{ .Values.cni.enabled | quote }}
- name: KUMA_RUNTIME_KUBERNETES_INJECTOR_SIDECAR_CONTAINER_IMAGE
  value: {{ include "kuma.formatImage" (dict "image" .Values.dataPlane.image "root" $) | quote }}
- name: KUMA_INJECTOR_INIT_CONTAINER_IMAGE
  value: {{ include "kuma.formatImage" (dict "image" .Values.dataPlane.initImage "root" $) | quote }}
{{- if .Values.dataPlane.dnsLogging }}
- name: KUMA_RUNTIME_KUBERNETES_INJECTOR_BUILTIN_DNS_LOGGING
  value: "true"
{{- end }}
{{- if and .Values.transparentProxy.configMap.enabled .Values.transparentProxy.configMap.config }}
- name: KUMA_RUNTIME_KUBERNETES_INJECTOR_TRANSPARENT_PROXY_CONFIGMAP_NAME
  value: {{ include "kuma.transparentProxyConfigMapName" . | quote }}
{{- end }}
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
- name: KUMA_RUNTIME_KUBERNETES_ALLOWED_USERS
  value: "system:serviceaccount:{{ .Release.Namespace }}:{{ include "kuma.name" . }}-control-plane"
{{- if .Values.experimental.sidecarContainers }}
- name: KUMA_EXPERIMENTAL_SIDECAR_CONTAINERS
  value: "true"
{{- end }}
{{- if .Values.cni.enabled }}
- name: KUMA_RUNTIME_KUBERNETES_NODE_TAINT_CONTROLLER_ENABLED
  value: "true"
- name: KUMA_RUNTIME_KUBERNETES_NODE_TAINT_CONTROLLER_CNI_APP
  value: "{{ include "kuma.name" . }}-cni"
- name: KUMA_RUNTIME_KUBERNETES_NODE_TAINT_CONTROLLER_CNI_NAMESPACE
  value: {{ .Values.cni.namespace }}
{{- end }}
{{- if .Values.experimental.ebpf.enabled }}
- name: KUMA_RUNTIME_KUBERNETES_INJECTOR_EBPF_ENABLED
  value: "true"
- name: KUMA_RUNTIME_KUBERNETES_INJECTOR_EBPF_INSTANCE_IP_ENV_VAR_NAME
  value: {{ .Values.experimental.ebpf.instanceIPEnvVarName }}
- name: KUMA_RUNTIME_KUBERNETES_INJECTOR_EBPF_BPFFS_PATH
  value: {{ .Values.experimental.ebpf.bpffsPath }}
- name: KUMA_RUNTIME_KUBERNETES_INJECTOR_EBPF_CGROUP_PATH
  value: {{ .Values.experimental.ebpf.cgroupPath }}
- name: KUMA_RUNTIME_KUBERNETES_INJECTOR_EBPF_TC_ATTACH_IFACE
  value: {{ .Values.experimental.ebpf.tcAttachIface }}
- name: KUMA_RUNTIME_KUBERNETES_INJECTOR_EBPF_PROGRAMS_SOURCE_PATH
  value: {{ .Values.experimental.ebpf.programsSourcePath }}
{{- end }}
{{- if .Values.controlPlane.tls.kdsZoneClient.skipVerify }}
- name: KUMA_MULTIZONE_ZONE_KDS_TLS_SKIP_VERIFY
  value: "true"
{{- end }}
- name: KUMA_PLUGIN_POLICIES_ENABLED
  value: {{ include "kuma.pluginPoliciesEnabled" . | quote }}
{{- if .Values.controlPlane.supportGatewaySecretsInAllNamespaces }}
- name: KUMA_RUNTIME_KUBERNETES_SUPPORT_GATEWAY_SECRETS_IN_ALL_NAMESPACES
  value: true
{{- end }}
{{- end }}

{{- define "kuma.controlPlane.tls.general.caSecretName" -}}
{{ .Values.controlPlane.tls.general.caSecretName | default .Values.controlPlane.tls.general.secretName | default (printf "%s-tls-cert" (include "kuma.name" .)) | quote }}
{{- end }}

{{- define "kuma.universal.defaultEnv" -}}
{{ if eq .Values.controlPlane.mode "zone" }}
  {{ if .Values.ingress.enabled }}
    {{ fail "Can't have ingress.enabled when running controlPlane.mode=='universal'" }}
  {{ end }}
  {{ if .Values.egress.enabled }}
    {{ fail "Can't have egress.enabled when running controlPlane.mode=='universal'" }}
  {{ end }}
{{ end }}

env:
- name: KUMA_PLUGIN_POLICIES_ENABLED
  value: {{ include "kuma.pluginPoliciesEnabled" . | quote }}
- name: KUMA_GENERAL_WORK_DIR
  value: "/tmp/kuma"
- name: KUMA_ENVIRONMENT
  value: "universal"
- name: KUMA_STORE_TYPE
  value: "postgres"
- name: KUMA_STORE_POSTGRES_PORT
  value: "{{ .Values.postgres.port }}"
- name: KUMA_DEFAULTS_SKIP_MESH_CREATION
  value: {{ .Values.controlPlane.defaults.skipMeshCreation | quote }}
{{ if and (eq .Values.controlPlane.mode "zone") .Values.controlPlane.tls.general.secretName }}
- name: KUMA_GENERAL_TLS_CERT_FILE
  value: /var/run/secrets/kuma.io/tls-cert/tls.crt
- name: KUMA_GENERAL_TLS_KEY_FILE
  value: /var/run/secrets/kuma.io/tls-cert/tls.key
{{ end }}
- name: KUMA_MODE
  value: {{ .Values.controlPlane.mode | quote }}
{{- if eq .Values.controlPlane.mode "zone" }}
- name: KUMA_MULTIZONE_ZONE_GLOBAL_ADDRESS
  value: {{ .Values.controlPlane.kdsGlobalAddress }}
{{- end }}
{{- if .Values.controlPlane.zone }}
- name: KUMA_MULTIZONE_ZONE_NAME
  value: {{ .Values.controlPlane.zone | quote }}
{{- end }}
{{- if and (eq .Values.controlPlane.mode "zone") (or .Values.controlPlane.tls.kdsZoneClient.secretName .Values.controlPlane.tls.kdsZoneClient.create) }}
- name: KUMA_MULTIZONE_ZONE_KDS_ROOT_CA_FILE
  value: /var/run/secrets/kuma.io/kds-client-tls-cert/ca.crt
{{- end }}
{{- if .Values.controlPlane.tls.kdsZoneClient.skipVerify }}
- name: KUMA_MULTIZONE_ZONE_KDS_TLS_SKIP_VERIFY
  value: "true"
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
- name: KUMA_STORE_POSTGRES_TLS_MODE
  value: {{ .Values.postgres.tls.mode }}
{{- if or (eq .Values.postgres.tls.mode "verifyCa") (eq .Values.postgres.tls.mode "verifyFull") }}
{{- if empty .Values.postgres.tls.caSecretName }}
{{ fail "if mode is 'verifyCa' or 'verifyFull' then you must provide .Values.postgres.tls.caSecretName" }}
{{- end }}
{{- if .Values.postgres.tls.secretName }}
- name: KUMA_STORE_POSTGRES_TLS_CERT_PATH
  value: /var/run/secrets/kuma.io/postgres-tls-cert/tls.crt
- name: KUMA_STORE_POSTGRES_TLS_KEY_PATH
  value: /var/run/secrets/kuma.io/postgres-tls-cert/tls.key
{{- end }}
{{- if .Values.postgres.tls.caSecretName }}
- name: KUMA_STORE_POSTGRES_TLS_CA_PATH
  value: /var/run/secrets/kuma.io/postgres-tls-cert/ca.crt
{{- end }}
{{- if .Values.postgres.tls.disableSSLSNI }}
- name: KUMA_STORE_POSTGRES_TLS_DISABLE_SSLSNI
  value: {{ .Values.postgres.tls.disableSSLSNI }}
{{- end }}
{{- end }}
{{- end }}
