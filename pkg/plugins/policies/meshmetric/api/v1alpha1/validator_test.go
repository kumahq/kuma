package v1alpha1

import (
	. "github.com/onsi/ginkgo/v2"

	"github.com/kumahq/kuma/v3/pkg/core/validators"
	. "github.com/kumahq/kuma/v3/pkg/test/resources/validators"
)

var _ = Describe("validation", func() {
	DescribeValidCases(
		NewMeshMetricResource,
		Entry("full spec", `
type: MeshMetric
mesh: mesh-1
name: metrics-1
targetRef:
  kind: Mesh
default:
  sidecar:
    profiles:
      appendProfiles:
        - name: None
        - name: Basic
      exclude:
        - type: Regex
          match: "my_match.*"
      include:
        - type: Prefix
          match: "my_match"
    includeUnused: true
  applications:
    - path: "metrics/prometheus"
      port: 8888
    - port: 8000
  backends:
    - type: Prometheus
      prometheus:
        clientId: main-backend 
        port: 5670
        path: /metrics
        tls:
          mode: "ProvidedTLS"
    - type: OpenTelemetry
      openTelemetry:
        endpoint: otel-collector:4778
`),
		Entry("openTelemetry with backendRef", `
type: MeshMetric
mesh: mesh-1
name: metrics-1
targetRef:
  kind: Mesh
default:
  backends:
    - type: OpenTelemetry
      openTelemetry:
        backendRef:
          kind: MeshOpenTelemetryBackend
          labels:
            kuma.io/display-name: my-otel
`),
	)

	DescribeErrorCases(
		NewMeshMetricResource,
		ErrorCase(
			"missing Prometheus config",
			validators.Violation{
				Field:   "spec.default.backends.backend[0].prometheus",
				Message: "must be defined",
			},
			`
type: MeshMetric
mesh: mesh-1
name: metrics-1
targetRef:
  kind: Mesh
default:
  backends:
    - type: Prometheus
`),
		ErrorCase(
			"invalid port for prometheus listener",
			validators.Violation{
				Field:   "spec.default.backends.backend[0].port",
				Message: "port must be a valid (1-65535)",
			},
			`
type: MeshMetric
mesh: mesh-1
name: metrics-1
targetRef:
  kind: Mesh
default:
  backends:
    - type: Prometheus
      prometheus:
        port: 95599
`),
		ErrorCase(
			"invalid port for application",
			validators.Violation{
				Field:   "spec.default.applications.application[0]",
				Message: "port must be a valid (1-65535)",
			},
			`
type: MeshMetric
mesh: mesh-1
name: metrics-1
targetRef:
  kind: Mesh
default:
  applications:
    - port: 95599
`),
		ErrorCase(
			"invalid exclude regexes",
			validators.Violation{
				Field:   "spec.default.sidecar.profiles.exclude[0].match",
				Message: "invalid regex",
			},
			`
type: MeshMetric
mesh: mesh-1
name: metrics-1
targetRef:
  kind: Mesh
default:
  sidecar:
    profiles:
      exclude:
        - type: Regex
          match: "())(!("
    includeUnused: true
`),
		ErrorCase(
			"invalid include types",
			validators.Violation{
				Field:   "spec.default.sidecar.profiles.include[0].type",
				Message: "unrecognized type 'not_supported' - 'Regex', 'Prefix', 'Exact' are supported",
			},
			`
type: MeshMetric
mesh: mesh-1
name: metrics-1
targetRef:
  kind: Mesh
default:
  sidecar:
    profiles:
      include:
        - type: not_supported
`),
		ErrorCase(
			"invalid profile",
			validators.Violation{
				Field:   "spec.default.sidecar.profiles.appendProfiles[0].name",
				Message: "unrecognized profile name 'not_supported' - 'All', 'None', 'Basic' are supported",
			},
			`
type: MeshMetric
mesh: mesh-1
name: metrics-1
targetRef:
  kind: Mesh
default:
  sidecar:
    profiles:
      appendProfiles:
        - name: not_supported
`),
		ErrorCase(
			"invalid endpoint missing port",
			validators.Violation{
				Field:   "spec.default.backends.backend[0].openTelemetry.endpoint",
				Message: "must be in host:port format",
			},
			`
type: MeshMetric
mesh: mesh-1
name: metrics-1
targetRef:
  kind: Mesh
default:
  backends:
    - type: OpenTelemetry
      openTelemetry:
        endpoint: "asdasd123"
`),
		ErrorCase(
			"invalid endpoint with URL scheme",
			validators.Violation{
				Field:   "spec.default.backends.backend[0].openTelemetry.endpoint",
				Message: "must be in host:port format, not a URL",
			},
			`
type: MeshMetric
mesh: mesh-1
name: metrics-1
targetRef:
  kind: Mesh
default:
  backends:
    - type: OpenTelemetry
      openTelemetry:
        endpoint: "http://endpoint:8023"
`),
		ErrorCase(
			"openTelemetry neither endpoint nor backendRef",
			validators.Violation{
				Field:   "spec.default.backends.backend[0].openTelemetry",
				Message: "openTelemetry must have exactly one defined: endpoint, backendRef",
			},
			`
type: MeshMetric
mesh: mesh-1
name: metrics-1
targetRef:
  kind: Mesh
default:
  backends:
    - type: OpenTelemetry
      openTelemetry: {}
`),
		ErrorCase(
			"openTelemetry both endpoint and backendRef",
			validators.Violation{
				Field:   "spec.default.backends.backend[0].openTelemetry",
				Message: "openTelemetry must have only one type defined: endpoint, backendRef",
			},
			`
type: MeshMetric
mesh: mesh-1
name: metrics-1
targetRef:
  kind: Mesh
default:
  backends:
    - type: OpenTelemetry
      openTelemetry:
        endpoint: otel-collector:4778
        backendRef:
          kind: MeshOpenTelemetryBackend
          labels:
            kuma.io/display-name: my-otel
`),
		ErrorCase(
			"openTelemetry backendRef no labels",
			validators.Violation{
				Field:   "spec.default.backends.backend[0].openTelemetry.backendRef",
				Message: "backendRef must have exactly one defined: labels",
			},
			`
type: MeshMetric
mesh: mesh-1
name: metrics-1
targetRef:
  kind: Mesh
default:
  backends:
    - type: OpenTelemetry
      openTelemetry:
        backendRef:
          kind: MeshOpenTelemetryBackend
`),
		ErrorCase(
			"undefined openTelemetry backend when type is OpenTelemetry",
			validators.Violation{
				Field:   "spec.default.backends.backend[0].openTelemetry",
				Message: "must be defined",
			},
			`
type: MeshMetric
mesh: mesh-1
name: metrics-1
targetRef:
  kind: Mesh
default:
  backends:
    - type: OpenTelemetry
      prometheus:
        port: 5670
`),
	)
})
