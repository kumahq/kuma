package mesh_test

import (
	. "github.com/onsi/ginkgo/v2"

	. "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/validators"
	_ "github.com/kumahq/kuma/pkg/plugins/runtime/gateway/register"
	. "github.com/kumahq/kuma/pkg/test/resources/validators"
)

var _ = Describe("Gateway", func() {
	DescribeValidCases(
		NewMeshGatewayResource,
		Entry("HTTPS listener", `
type: MeshGateway
name: gateway
mesh: default
selectors:
  - match:
      kuma.io/service: gateway
tags:
  product: edge
conf:
  listeners:
  - hostname: www-1.example.com
    port: 443
    protocol: HTTP
    tags:
      name: https`,
		),
		Entry("HTTPS listener without tags", `
type: MeshGateway
name: gateway
mesh: default
selectors:
  - match:
      kuma.io/service: gateway
tags:
  product: edge
conf:
  listeners:
  - hostname: www-1.example.com
    port: 443
    protocol: HTTP`,
		),
		Entry("crossMesh, HTTP, with hostname", `
type: MeshGateway
name: gateway
mesh: default
selectors:
  - match:
      kuma.io/service: gateway
conf:
  listeners:
  - hostname: www-1.example.com
    port: 443
    crossMesh: true
    protocol: HTTP`,
		),
		Entry("crossMesh with no hostname", `
type: MeshGateway
name: gateway
mesh: default
selectors:
  - match:
      kuma.io/service: gateway
tags:
  product: edge
conf:
  listeners:
  - protocol: HTTP
    port: 99
    crossMesh: true
    tags:
      name: http`,
		),
		Entry("listeners with connectionLimits", `
type: MeshGateway
name: gateway
mesh: default
selectors:
  - match:
      kuma.io/service: gateway
tags:
  product: edge
conf:
  listeners:
  - protocol: HTTP
    hostname: one.com
    port: 99
    resources:
      connectionLimit: 2
  - protocol: HTTP
    hostname: two.com
    port: 99
    resources:
      connectionLimit: 2
`,
		),
		Entry("TLS listener with TERMINATE", `
type: MeshGateway
name: gateway
mesh: default
selectors:
  - match:
      kuma.io/service: gateway
tags:
  product: edge
conf:
  listeners:
  - port: 443
    protocol: TLS
    tls:
      mode: TERMINATE
      certificates:
      - secret: example-kuma-io-certificate
`,
		),
	)

	DescribeErrorCases(
		NewMeshGatewayResource,
		ErrorCases("doesn't have any selectors",
			[]validators.Violation{{
				Field:   `selectors`,
				Message: `must have at least one element`,
			}, {
				Field:   "conf.listeners[0].tls.certificates",
				Message: "cannot be empty in TLS termination mode",
			}}, `
type: MeshGateway
name: gateway
mesh: default
selectors:
tags:
  product: edge
conf:
  listeners:
  - port: 443
    protocol: HTTPS
    tls:
      mode: TERMINATE
    tags:
      name: https
`),

		ErrorCase("has a service tag",
			validators.Violation{
				Field:   `tags["kuma.io/service"]`,
				Message: `tag name must not be "kuma.io/service"`,
			}, `
type: MeshGateway
name: gateway
mesh: default
selectors:
  - match:
      kuma.io/service: gateway
tags:
  product: edge
  kuma.io/service: gateway
conf:
  listeners:
  - port: 443
    protocol: HTTP
    tags:
      name: https
`),

		ErrorCase("doesn't have a configuration spec",
			validators.Violation{
				Field:   "conf",
				Message: "cannot be empty",
			}, `
type: MeshGateway
name: gateway
mesh: default
selectors:
  - match:
      kuma.io/service: gateway
tags:
  product: edge
conf:
`),

		ErrorCase("has an invalid port",
			validators.Violation{
				Field:   "conf.listeners[0].port",
				Message: "port must be in the range [1, 65535]",
			}, `
type: MeshGateway
name: gateway
mesh: default
selectors:
  - match:
      kuma.io/service: gateway
tags:
  product: edge
conf:
  listeners:
  - protocol: HTTP
    tags:
      name: https
`),

		ErrorCase("has an empty protocol",
			validators.Violation{
				Field:   "conf.listeners[0].protocol",
				Message: "cannot be empty",
			}, `
type: MeshGateway
name: gateway
mesh: default
selectors:
  - match:
      kuma.io/service: gateway
tags:
  product: edge
conf:
  listeners:
  - port: 99
    tags:
      name: https
`),

		ErrorCase("has an empty TLS mode",
			validators.Violation{
				Field:   "conf.listeners[0].tls.mode",
				Message: "cannot be empty",
			}, `
type: MeshGateway
name: gateway
mesh: default
selectors:
  - match:
      kuma.io/service: gateway
tags:
  product: edge
conf:
  listeners:
  - protocol: HTTPS
    port: 99
    tags:
      name: https
    tls:
      options:
`),

		ErrorCase("is missing a TLS termination secret",
			validators.Violation{
				Field:   "conf.listeners[0].tls.certificates",
				Message: "cannot be empty in TLS termination mode",
			}, `
type: MeshGateway
name: gateway
mesh: default
selectors:
  - match:
      kuma.io/service: gateway
tags:
  product: edge
conf:
  listeners:
  - protocol: HTTPS
    port: 99
    tags:
      name: https
    tls:
      mode: TERMINATE
`),

		ErrorCase("has an invalid wildcard",
			validators.Violation{
				Field:   "conf.listeners[0].hostname",
				Message: "invalid wildcard domain",
			}, `
type: MeshGateway
name: gateway
mesh: default
selectors:
  - match:
      kuma.io/service: gateway
tags:
  product: edge
conf:
  listeners:
  - hostname: "*.foo.*.example.com"
    protocol: HTTP
    port: 99
    tags:
      name: https
`),

		ErrorCases("has an invalid hostname",
			[]validators.Violation{{
				Field:   "conf.listeners[0].hostname",
				Message: "invalid hostname",
			}, {
				Field:   "conf.listeners[1].hostname",
				Message: "must be at most 253 characters",
			}}, `
type: MeshGateway
name: gateway
mesh: default
selectors:
  - match:
      kuma.io/service: gateway
tags:
  product: edge
conf:
  listeners:
  - hostname: "foo.example$.com"
    protocol: HTTP
    port: 99
    tags:
      name: https
  - hostname: "ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff.com"
    protocol: HTTP
    port: 100
    tags:
      name: https
`),

		ErrorCase("crossMesh and HTTPS",
			validators.Violation{
				Field:   "conf.listeners[0].protocol",
				Message: "protocol is not supported with crossMesh",
			}, `
type: MeshGateway
name: gateway
mesh: default
selectors:
  - match:
      kuma.io/service: gateway
tags:
  product: edge
conf:
  listeners:
  - hostname: "foo.example.com"
    protocol: HTTPS
    port: 99
    crossMesh: true
    tags:
      name: https
`),

		ErrorCase("HTTPS and PASSTHROUGH",
			validators.Violation{
				Field:   "conf.listeners[0].tls.mode",
				Message: "mode is not supported on HTTPS listeners",
			}, `
type: MeshGateway
name: gateway
mesh: default
selectors:
  - match:
      kuma.io/service: gateway
tags:
  product: edge
conf:
  listeners:
  - hostname: "foo.example.com"
    protocol: HTTPS
    port: 99
    tags:
      name: https
    tls:
      mode: PASSTHROUGH
`),

		ErrorCase("crossMesh and multiple services",
			validators.Violation{
				Field:   "selectors[1]",
				Message: "there can be at most one selector",
			}, `
type: MeshGateway
name: gateway
mesh: default
selectors:
  - match:
      kuma.io/service: gateway
  - match:
      kuma.io/service: other-gateway
conf:
  listeners:
  - hostname: www-1.example.com
    port: 443
    crossMesh: true
    protocol: HTTP
`),

		ErrorCases("protocol conflict",
			[]validators.Violation{{
				Field:   "conf.listeners[0]",
				Message: "protocol conflicts with other listeners on this port",
			}, {
				Field:   "conf.listeners[1]",
				Message: "protocol conflicts with other listeners on this port",
			}, {
				Field:   "conf.listeners[1].tls.certificates",
				Message: "cannot be empty in TLS termination mode",
			}}, `
type: MeshGateway
name: gateway
mesh: default
selectors:
  - match:
      kuma.io/service: gateway
conf:
  listeners:
  - hostname: www-1.example.com
    port: 443
    protocol: TCP
  - hostname: www-2.example.com
    port: 443
    protocol: HTTPS
    tls:
      mode: TERMINATE
`),

		ErrorCases("hostname conflict",
			[]validators.Violation{{
				Field:   "conf.listeners[0]",
				Message: "multiple listeners for hostname on this port",
			}, {
				Field:   "conf.listeners[1]",
				Message: "multiple listeners for hostname on this port",
			}}, `
type: MeshGateway
name: gateway
mesh: default
selectors:
  - match:
      kuma.io/service: gateway
conf:
  listeners:
  - hostname: www-1.example.com
    port: 443
    protocol: TCP
  - hostname: www-1.example.com
    port: 443
    protocol: TCP
`),

		ErrorCases("hostname, protocol and resource conflict",
			[]validators.Violation{{
				Field:   "conf.listeners[0]",
				Message: "protocol conflicts with other listeners on this port",
			}, {
				Field:   "conf.listeners[1]",
				Message: "protocol conflicts with other listeners on this port",
			}, {
				Field:   "conf.listeners[0]",
				Message: "multiple listeners for hostname on this port",
			}, {
				Field:   "conf.listeners[1]",
				Message: "multiple listeners for hostname on this port",
			}, {
				Field:   "conf.listeners[1].tls.certificates",
				Message: "cannot be empty in TLS termination mode",
			}, {
				Field:   "conf.listeners[0].resources.connectionLimit",
				Message: "conflicting values for this port",
			}, {
				Field:   "conf.listeners[1].resources.connectionLimit",
				Message: "conflicting values for this port",
			}}, `
type: MeshGateway
name: gateway
mesh: default
selectors:
  - match:
      kuma.io/service: gateway
conf:
  listeners:
  - hostname: www-1.example.com
    port: 443
    protocol: TCP
    resources:
      connectionLimit: 2
  - hostname: www-1.example.com
    port: 443
    protocol: HTTPS
    tls:
      mode: TERMINATE
    resources:
      connectionLimit: 1
`),
	)
})
