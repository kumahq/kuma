package v1alpha1_test

import (
	. "github.com/onsi/ginkgo/v2"

	"github.com/kumahq/kuma/pkg/core/validators"
	meshfaultinjection_proto "github.com/kumahq/kuma/pkg/plugins/policies/meshfaultinjection/api/v1alpha1"
	. "github.com/kumahq/kuma/pkg/test/resources/validators"
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
from:
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
            limit: 100Mbps
            percentage: 5
        - abort:
            httpStatus: 500
            percentage: "50.5"
`),
		Entry("empty faults", `
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
      http: []
`),
		Entry("Kind Mesh with to and only gateway", `
type: MeshFaultInjection
mesh: mesh-1
name: fi1
targetRef:
  kind: Mesh
  proxyTypes: ["Gateway"]
to:
  - targetRef:
      kind: Mesh
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
					Message: `must be in inclusive range [100, 599]`,
				},
				{
					Field:   `spec.from[0].default.http.abort[0].percentage`,
					Message: `must be in inclusive range [0.0, 100.0]`,
				},
				{
					Field:   "spec.from[0].default.http.delay[1].value",
					Message: "must not be negative when defined",
				},
				{
					Field:   `spec.from[0].default.http.delay[1].percentage`,
					Message: `must be in inclusive range [0.0, 100.0]`,
				},
				{
					Field:   `spec.from[0].default.http.responseBandwidth[2].responseBandwidth`,
					Message: `must be in kbps/Mbps/Gbps units`,
				},
				{
					Field:   `spec.from[0].default.http.responseBandwidth[2].percentage`,
					Message: `must be in inclusive range [0.0, 100.0]`,
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
					Message: "must be in inclusive range [100, 599]",
				},
				{
					Field:   "spec.from[0].default.http.responseBandwidth[2].responseBandwidth",
					Message: "must be in kbps/Mbps/Gbps units",
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
		ErrorCases("sectionName with outbound policy",
			[]validators.Violation{
				{
					Field:   "spec.targetRef.sectionName",
					Message: "can only be used with inbound policies",
				},
				{
					Field:   "spec.to",
					Message: "must not be defined",
				},
			}, `
type: MeshFaultInjection
mesh: mesh-1
name: fi1
targetRef:
  kind: Dataplane
  sectionName: test
to:
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
		ErrorCases("incorrect value in percentage",
			[]validators.Violation{
				{
					Field:   "spec.from[0].default.http.responseBandwidth[0].responseBandwidth",
					Message: "must be in kbps/Mbps/Gbps units",
				},
				{
					Field:   "spec.from[0].default.http.responseBandwidth[0].percentage",
					Message: "string must be a valid number",
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
      - responseBandwidth:
          limit: 1000
          percentage: "xyz"`,
		),
		ErrorCases("MeshGateway and targetRefs",
			[]validators.Violation{
				{
					Field:   "spec.from",
					Message: "must not be defined when the scope includes a Gateway, exclude non-gateway resources or select only gateways and use spec.to",
				},
				{
					Field:   "spec.to[1].targetRef.kind",
					Message: "value is not supported",
				},
			}, `
type: MeshFaultInjection
mesh: mesh-1
name: fi1
targetRef:
  kind: MeshGateway
  name: edge
to:
  - targetRef:
      kind: Mesh
    default:
      http:
      - abort:
          httpStatus: 503
          percentage: 50
  - targetRef:
      kind: MeshService
      name: backend
    default:
      http:
      - abort:
          httpStatus: 503
          percentage: 50
from:
  - targetRef:
      kind: Mesh
    default:
      http:
      - abort:
          httpStatus: 503
          percentage: 50`,
		),
	)
})
