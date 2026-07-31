package framework

import (
	"bytes"
	"text/template"

	"github.com/pkg/errors"
)

// DataplaneTemplateData represents the data for dataplane templates
type DataplaneTemplateData struct {
	// Core fields
	Name        string
	Mesh        string
	ServiceName string

	// Inbound configuration
	InboundPort        string
	InboundServicePort string
	ServiceAddress     string

	// Tags
	Protocol       string
	Version        string
	Instance       string
	Team           string
	AdditionalTags map[string]string
	Labels         map[string]string

	// Service probe
	ServiceProbe bool

	// Outbound configuration
	Outbounds []OutboundConfig

	// Transparent proxy configuration
	TransparentProxy *TransparentProxyConfig

	// Listener configuration (zone proxy mode). When set, renders a listeners
	// block instead of inbound/outbound.
	Listeners []ListenerConfig

	// Additional raw YAML to append
	AppendConfig string
}

// OutboundConfig represents an outbound configuration
type OutboundConfig struct {
	Port    string
	Service string
}

// TransparentProxyConfig represents transparent proxy configuration
type TransparentProxyConfig struct {
	RedirectPortInbound  string
	RedirectPortOutbound string
	// ReachableBackends is the raw YAML body rendered under
	// networking.transparentProxying.reachableBackends (e.g. a `refs:` list).
	ReachableBackends string
}

// ListenerConfig represents a zone proxy listener (ZoneIngress or ZoneEgress).
type ListenerConfig struct {
	Type string // "ZoneIngress" or "ZoneEgress"
	Name string
	Port int
}

var dataplaneTemplate = template.Must(template.New("dataplane").Parse(`
type: Dataplane
mesh: {{ .Mesh }}
name: {{ "{{ name }}" }}
{{- if .Labels }}
labels:
{{- range $key, $value := .Labels }}
  {{ $key }}: {{ $value }}
{{- end }}
{{- end }}
networking:
  address: {{ "{{ address }}" }}
{{- if .Listeners }}
  listeners:
{{- range .Listeners }}
  - type: {{ .Type }}
    address: {{ "{{ address }}" }}
    port: {{ .Port }}
    name: {{ .Name }}
{{- end }}
{{- else }}
  inbound:
  - port: {{ .InboundPort }}
{{- if .InboundServicePort }}
    servicePort: {{ .InboundServicePort }}
{{- end }}
{{- if .ServiceAddress }}
    serviceAddress: {{ .ServiceAddress }}
{{- end }}
{{- if .ServiceProbe }}
    serviceProbe:
      tcp: {}
{{- end }}
{{- if .Protocol }}
    protocol: {{ .Protocol }}
{{- end }}
    tags:
      kuma.io/service: {{ .ServiceName }}
{{- if .Protocol }}
      kuma.io/protocol: {{ .Protocol }}
{{- end }}
{{- if .Team }}
      team: {{ .Team }}
{{- end }}
{{- if .Version }}
      version: {{ .Version }}
{{- end }}
{{- if .Instance }}
      instance: '{{ .Instance }}'
{{- end }}
{{- range $key, $value := .AdditionalTags }}
      {{ $key }}: {{ $value }}
{{- end }}
{{- if .Outbounds }}
  outbound:
{{- range .Outbounds }}
  - port: {{ .Port }}
    tags:
      kuma.io/service: {{ .Service }}
{{- end }}
{{- end }}
{{- if .TransparentProxy }}
  transparentProxying:
    redirectPortInbound: {{ .TransparentProxy.RedirectPortInbound }}
    redirectPortOutbound: {{ .TransparentProxy.RedirectPortOutbound }}
{{- if .TransparentProxy.ReachableBackends }}
    reachableBackends:
{{ .TransparentProxy.ReachableBackends }}
{{- end }}
{{- end }}
{{- end }}
{{- if .AppendConfig }}
{{ .AppendConfig }}
{{- end }}`))

// RenderDataplaneTemplate renders a dataplane template with the given data.
// When Listeners is set, renders a zone proxy dataplane with a listeners block
// instead of inbound/outbound.
func RenderDataplaneTemplate(data DataplaneTemplateData) (string, error) {
	var buf bytes.Buffer
	if err := dataplaneTemplate.Execute(&buf, data); err != nil {
		return "", errors.Wrap(err, "failed to execute dataplane template")
	}
	return buf.String(), nil
}
