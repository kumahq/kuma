{{/* vim: set filetype=mustache: */}}

{{/*
Generate the resource name for a per-mesh zone proxy component.
params: { root: $, meshName: string, role: string }
role is one of: ingress, egress, zoneproxy (combined)
*/}}
{{- define "kuma.mesh.zoneproxy.name" -}}
{{- printf "%s-%s-%s" (include "kuma.name" .root) .meshName .role -}}
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
