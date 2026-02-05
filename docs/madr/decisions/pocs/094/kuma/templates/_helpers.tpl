{{/* Minimal helpers — the parent chart would normally have more. */}}
{{- define "kuma.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}
