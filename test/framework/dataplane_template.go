package framework

import (
	"bytes"
	"strings"
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
	ReachableServices    []string
}

// ZoneIngressTemplateData represents zone ingress template data
type ZoneIngressTemplateData struct {
	Name              string
	AdvertisedAddress string
	AdvertisedPort    int
	Port              int
}

// ZoneEgressTemplateData represents zone egress template data
type ZoneEgressTemplateData struct {
	Name string
	Port int
}

var (
	dataplaneTemplate = template.Must(template.New("dataplane").Funcs(template.FuncMap{
		"joinStrings": strings.Join,
	}).Parse(`
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
{{- if .TransparentProxy.ReachableServices }}
    reachableServices: [{{ joinStrings .TransparentProxy.ReachableServices "," }}]
{{- end }}
{{- end }}
{{- if .AppendConfig }}
{{ .AppendConfig }}
{{- end }}`))
	zoneIngressTemplate = template.Must(template.New("zoneingress").Parse(`
type: ZoneIngress
name: {{ .Name }}
networking:
  address: {{ "{{ address }}" }}
  advertisedAddress: {{ .AdvertisedAddress }}
  advertisedPort: {{ .AdvertisedPort }}
  port: {{ .Port }}`))
	zoneEgressTemplate = template.Must(template.New("zoneegress").Parse(`
type: ZoneEgress
name: egress
networking:
  address: {{ "{{ address }}" }}
  port: {{ .Port }}`))
)

// RenderDataplaneTemplate renders a dataplane template with the given data
func RenderDataplaneTemplate(data DataplaneTemplateData) (string, error) {
	var buf bytes.Buffer
	if err := dataplaneTemplate.Execute(&buf, data); err != nil {
		return "", errors.Wrap(err, "failed to execute dataplane template")
	}
	return buf.String(), nil
}

// RenderZoneIngressTemplate renders a zone ingress template with the given data
func RenderZoneIngressTemplate(data ZoneIngressTemplateData) (string, error) {
	var buf bytes.Buffer
	if err := zoneIngressTemplate.Execute(&buf, data); err != nil {
		return "", errors.Wrap(err, "failed to execute zone ingress template")
	}
	return buf.String(), nil
}

// RenderZoneEgressTemplate renders a zone egress template with the given data
func RenderZoneEgressTemplate(data ZoneEgressTemplateData) (string, error) {
	var buf bytes.Buffer
	if err := zoneEgressTemplate.Execute(&buf, data); err != nil {
		return "", errors.Wrap(err, "failed to execute zone egress template")
	}
	return buf.String(), nil
}
