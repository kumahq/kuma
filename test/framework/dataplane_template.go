package framework

import (
	"bytes"
	"maps"
	"text/template"

	"github.com/pkg/errors"

	mesh_proto "github.com/kumahq/kuma/v3/api/mesh/v1alpha1"
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

// OutboundConfig represents an outbound configuration. The destination is
// selected through a MeshService backendRef, so an outbound whose MeshService
// does not exist resolves to nothing instead of producing a listener for a
// service that is not there. Service is matched on kuma.io/display-name.
type OutboundConfig struct {
	Port        string
	Service     string
	ServicePort string
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
  {{ $key }}: "{{ $value }}"
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
{{- if .Outbounds }}
  outbound:
{{- range .Outbounds }}
  - port: {{ .Port }}
    backendRef:
      kind: MeshService
      labels:
        kuma.io/display-name: {{ .Service }}
      port: {{ .ServicePort }}
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
// instead of inbound/outbound. ServiceName/Team/Version/Instance/AdditionalTags
// are rendered as Dataplane labels rather than inbound tags: service identity
// and endpoint load-balancing identity (envoy.lb metadata) are now sourced
// solely from Dataplane labels, not inbound tags (see pkg/xds/topology/outbound.go).
func RenderDataplaneTemplate(data DataplaneTemplateData) (string, error) {
	labels := make(map[string]string, len(data.Labels)+len(data.AdditionalTags)+3)
	maps.Copy(labels, data.Labels)
	maps.Copy(labels, data.AdditionalTags)
	if data.ServiceName != "" {
		labels[mesh_proto.ServiceTag] = data.ServiceName
	}
	if data.Team != "" {
		labels["team"] = data.Team
	}
	if data.Version != "" {
		labels["version"] = data.Version
	}
	if data.Instance != "" {
		labels["instance"] = data.Instance
	}
	data.Labels = labels

	var buf bytes.Buffer
	if err := dataplaneTemplate.Execute(&buf, data); err != nil {
		return "", errors.Wrap(err, "failed to execute dataplane template")
	}
	return buf.String(), nil
}
