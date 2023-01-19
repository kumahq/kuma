package v1alpha1_test

import (
	. "github.com/onsi/ginkgo/v2"

	"github.com/kumahq/kuma/pkg/core/validators"
	meshfaultinjection_proto "github.com/kumahq/kuma/pkg/plugins/policies/meshfaultinjection/api/v1alpha1"
	. "github.com/kumahq/kuma/pkg/test/resources"
)

var _ = Describe("MeshFaultInjection", func() {
	DescribeValidCases(
		meshfaultinjection_proto.NewMeshFaultInjectionResource,
		Entry("accepts valid resource", `
type: MeshFaultInjection
mesh: mesh-1
name: fi1
targetRef:
  kind: MeshService
  name: backend
to:
  - targetRef:
      kind: MeshService
      name: web-backend
    default:
      http:
        - abort:
            httpStatus: 503
            percentage: 50
          delay:
            value: 10s
            percentage: 5
        - delay:
            value: 5s
            percentage: 5
        - responseBandwidth:
            limit: 100mbps
            percentage: 5
        - abort:
            httpStatus: 500
            percentage: 50
`),
		Entry("empty faults", `
type: MeshFaultInjection
mesh: mesh-1
name: fi1
targetRef:
  kind: MeshService
  name: backend
to:
  - targetRef:
      kind: MeshService
      name: web-backend
    default:
      http: []
`),
	)

	DescribeErrorCases(
		meshfaultinjection_proto.NewMeshFaultInjectionResource,
		ErrorCases("incorrect values",
			[]validators.Violation{
				{
					Field:   `spec.from[0].default.http.abort[0].httpStatus`,
					Message: `must be in range [100, 600)`,
				},
				{
					Field:   `spec.from[0].default.http.abort[0].percentage`,
					Message: `has to be in [0 - 100] range`,
				},
				{
					Field:   "spec.from[0].default.http.delay[1].value",
					Message: "must not be negative when defined",
				},
				{
					Field:   `spec.from[0].default.http.delay[1].percentage`,
					Message: `has to be in [0 - 100] range`,
				},
				{
					Field:   `spec.from[0].default.http.responseBandwidth[2].responseBandwidth`,
					Message: `has to be in kbps/mbps/gbps units`,
				},
				{
					Field:   `spec.from[0].default.http.responseBandwidth[2].percentage`,
					Message: `has to be in [0 - 100] range`,
				},
			}, `
type: MeshFaultInjection
mesh: mesh-1
name: fi1
targetRef:
  kind: MeshService
  name: backend
from:
  - targetRef:
      kind: MeshService
      name: web-backend
    default:
      http:
      - abort:
          httpStatus: 677
          percentage: 111
      - delay: 
          value: -5s
          percentage: 1111
      - responseBandwidth:
          limit: 1000
          percentage: 1111
`),
		ErrorCases("empty values",
			[]validators.Violation{
				{
					Field:   "spec.from[0].default.http.abort[0].httpStatus",
					Message: "must be in range [100, 600)",
				},
				{
					Field:   "spec.from[0].default.http.responseBandwidth[2].responseBandwidth",
					Message: "has to be in kbps/mbps/gbps units",
				},
			}, `
type: MeshFaultInjection
mesh: mesh-1
name: fi1
targetRef:
  kind: MeshService
  name: backend
from:
  - targetRef:
      kind: MeshService
      name: web-backend
    default:
      http:
      - abort: {}
      - delay: {}
      - responseBandwidth:
          limit: 1000
`),
	)
})
