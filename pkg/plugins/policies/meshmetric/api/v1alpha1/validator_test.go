package v1alpha1

import (
	"github.com/kumahq/kuma/pkg/core/validators"
	"github.com/kumahq/kuma/pkg/test/resources"
	. "github.com/onsi/ginkgo/v2"
)

var _ = Describe("validation", func() {
	resources.DescribeValidCases(
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
      name: main-backend
      prometheus:
        port: 5670
        path: /metrics
        tls:
          mode: "ProvidedTLS"
`),
	)

	resources.DescribeErrorCases(
		NewMeshMetricResource,
		resources.ErrorCase(
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
      name: main-backend
`),
		resources.ErrorCase(
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
	)
})
