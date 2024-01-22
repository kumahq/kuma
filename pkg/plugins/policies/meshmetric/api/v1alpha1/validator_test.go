package v1alpha1

import (
	. "github.com/onsi/ginkgo/v2"

	"github.com/kumahq/kuma/pkg/core/validators"
	. "github.com/kumahq/kuma/pkg/test/resources/validators"
)

var _ = Describe("validation", func() {
	DescribeValidCases(
		NewMeshMetricResource,
		Entry("full spec", `
type: MeshMetric
mesh: mesh-1
name: metrics-1
targetRef:
  kind: MeshService
  name: svc-1
default:
  sidecar:
    regex: "http2_.*"
    usedOnly: true
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
  kind: MeshService
  name: svc-1
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
  kind: MeshService
  name: svc-1
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
  kind: MeshService
  name: svc-1
default:
  applications:
    - port: 95599
`),
		ErrorCase(
			"invalid regex",
			validators.Violation{
				Field:   "spec.default.sidecar.regex",
				Message: "invalid regex",
			},
			`
type: MeshMetric
mesh: mesh-1
name: metrics-1
targetRef:
  kind: MeshService
  name: svc-1
default:
  sidecar:
    regex: "())(!("
    usedOnly: true
`),
		ErrorCase(
			"invalid url",
			validators.Violation{
				Field:   "spec.default.backends.backend[0].openTelemetry.endpoint",
				Message: "must be a valid url",
			},
			`
type: MeshMetric
mesh: mesh-1
name: metrics-1
targetRef:
  kind: MeshService
  name: svc-1
default:
  backends:
    - type: OpenTelemetry
      openTelemetry:
        endpoint: "asdasd123"
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
  kind: MeshService
  name: svc-1
default:
  backends:
    - type: OpenTelemetry
      prometheus:
        port: 5670
`),
	)
})
