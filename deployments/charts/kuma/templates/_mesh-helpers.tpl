{{/* vim: set filetype=mustache: */}}

{{/*
Generate the resource name for a per-mesh zone proxy component.
Truncated to 63 chars (Kubernetes label value limit).
params: { root: $, meshName: string, role: string }
role is one of: ingress, egress
*/}}
{{- define "kuma.mesh.zoneproxy.name" -}}
{{- printf "%s-%s-%s" (include "kuma.name" .root) .meshName .role | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Generate the Service name for a per-mesh zone proxy component.
Validates that the result does not exceed 63 characters.
params: { root: $, meshName: string, role: string, nameOverride: string }
*/}}
{{- define "kuma.mesh.zoneproxy.serviceName" -}}
{{- $default := include "kuma.mesh.zoneproxy.name" . -}}
{{- $name := default $default .nameOverride -}}
{{- if gt (len $name) 63 -}}
{{- fail (printf "zone proxy service name %q exceeds 63 characters; use service.name to set a shorter name" $name) -}}
{{- end -}}
{{- $name -}}
{{- end -}}

{{/*
Labels for per-mesh zone proxy resources.
params: { root: $, meshName: string, role: string }
*/}}
{{- define "kuma.mesh.zoneproxy.labels" -}}
app: {{ include "kuma.mesh.zoneproxy.name" . }}
kuma.io/mesh: {{ .meshName }}
{{ include "kuma.labels" .root }}
{{- end -}}

{{/*
Selector labels for per-mesh zone proxy resources.
params: { root: $, meshName: string, role: string }
*/}}
{{- define "kuma.mesh.zoneproxy.selectorLabels" -}}
app: {{ include "kuma.mesh.zoneproxy.name" . }}
{{ include "kuma.selectorLabels" .root }}
{{- end -}}

{{/*
Resolve the pause container image for a per-mesh zone proxy component.
Per-field override on meshes[].ingress.image / meshes[].egress.image falls
back to the chart-level default at .Values.zoneProxyImage so users enabling
a zone proxy on the default mesh do not need to re-specify registry/repo/tag.
params: { root: $, cfg: <meshes[].ingress or meshes[].egress>, meshName: string, role: string }
*/}}
{{- define "kuma.mesh.zoneproxy.image" -}}
{{- $default := .root.Values.zoneProxyImage | default dict -}}
{{- $override := and .cfg .cfg.image | default dict -}}
{{- $registry := default $default.registry $override.registry -}}
{{- $repository := default $default.repository $override.repository -}}
{{- $tag := default $default.tag $override.tag -}}
{{- if or (not $registry) (not $repository) (not $tag) -}}
{{- fail (printf "meshes[%s].%s.image is missing: set .Values.zoneProxyImage or meshes[].%s.image (registry/repository/tag)" .meshName .role .role) -}}
{{- end -}}
{{- printf "%s/%s:%s" $registry $repository $tag -}}
{{- end -}}
