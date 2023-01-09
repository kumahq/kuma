package v1alpha1

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"sigs.k8s.io/yaml"

	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	_ "github.com/kumahq/kuma/pkg/xds/envoy"
)

var _ = Describe("MeshProxyPatch", func() {
	Describe("Validate()", func() {
		DescribeTable("should pass validation",
			func(resYAML string) {
				// given
				resource := NewMeshProxyPatchResource()

				// when
				err := core_model.FromYAML([]byte(resYAML), &resource.Spec)
				Expect(err).ToNot(HaveOccurred())
				verr := resource.Validate()

				// then
				Expect(verr).To(BeNil())
			},
			Entry("cluster modifications", `
targetRef:
  kind: MeshService
  name: web-frontend
default:
  appendModifications:
  - cluster:
      operation: add
      value: |
        name: xyz
        connectTimeout: 5s
        type: STATIC
  - cluster:
      operation: patch
      value: |
        connectTimeout: 5s
  - cluster:
      operation: patch
      match:
        name: inbound:127.0.0.1:8080
      value: |
        connectTimeout: 5s
  - cluster:
      operation: patch
      match:
        origin: inbound
      value: |
        connectTimeout: 5s
  - cluster:
      operation: remove
  - cluster:
      operation: remove
      match:
        name: inbound:127.0.0.1:8080
  - cluster:
      operation: remove
      match:
        origin: inbound
    `),
			Entry("listener modifications", `
targetRef:
  kind: MeshServiceSubset
  name: backend
  tags:
    version: v2
default:
  appendModifications:
  - listener:
      operation: add
      value: |
        name: xyz
        address:
          socketAddress:
            address: 192.168.0.1
            portValue: 8080
  - listener:
      operation: patch
      value: |
        address:
          socketAddress:
            portValue: 8080
  - listener:
      operation: patch
      match:
        name: inbound:127.0.0.1:8080
      value: |
        address:
          socketAddress:
            portValue: 8080
  - listener:
      operation: patch
      match:
        origin: inbound
      value: |
        address:
          socketAddress:
            portValue: 8080
  - listener:
      operation: remove
  - listener:
      operation: remove
      match:
        name: inbound:127.0.0.1:8080
  - listener:
      operation: remove
      match:
        origin: inbound
    `),
			Entry("network filter modifications", `
targetRef:
  kind: Mesh
default:
  appendModifications:
  - networkFilter:
      operation: addFirst
      value: |
        name: envoy.filters.network.tcp_proxy
        typedConfig:
          '@type': type.googleapis.com/envoy.extensions.filters.network.tcp_proxy.v3.TcpProxy
          cluster: backend
  - networkFilter:
      operation: addLast
      value: |
        name: envoy.filters.network.tcp_proxy
        typedConfig:
          '@type': type.googleapis.com/envoy.extensions.filters.network.tcp_proxy.v3.TcpProxy
          cluster: backend
  - networkFilter:
      operation: addBefore
      match:
        name: envoy.filters.network.direct_response
      value: |
        name: envoy.filters.network.tcp_proxy
        typedConfig:
          '@type': type.googleapis.com/envoy.extensions.filters.network.tcp_proxy.v3.TcpProxy
          cluster: backend
  - networkFilter:
      operation: addAfter
      match:
        name: envoy.filters.network.direct_response
        listenerName: inbound:127.0.0.0:8080
      value: |
        name: envoy.filters.network.tcp_proxy
        typedConfig:
          '@type': type.googleapis.com/envoy.extensions.filters.network.tcp_proxy.v3.TcpProxy
          cluster: backend
  - networkFilter:
      operation: patch
      match:
        name: envoy.filters.network.tcp_proxy
        listenerName: inbound:127.0.0.0:8080
      value: |
        name: envoy.filters.network.tcp_proxy
        typedConfig:
          '@type': type.googleapis.com/envoy.extensions.filters.network.tcp_proxy.v3.TcpProxy
          cluster: backend
  - networkFilter:
      operation: remove
    `),
			Entry("http filter modifications", `
targetRef:
  kind: MeshSubset
  tags:
    kuma.io/zone: east
default:
  appendModifications:
  - httpFilter:
      operation: addFirst
      value: |
        name: envoy.filters.http.router
        typedConfig:
          '@type': type.googleapis.com/envoy.extensions.filters.http.router.v3.Router
          dynamicStats: false
  - httpFilter:
      operation: addLast
      value: |
        name: envoy.filters.http.router
        typedConfig:
          '@type': type.googleapis.com/envoy.extensions.filters.http.router.v3.Router
          dynamicStats: false
  - httpFilter:
      operation: addAfter
      match:
        name: envoy.filters.http.router
      value: |
        name: envoy.filters.http.gzip
  - httpFilter:
      operation: addAfter
      match:
        name: envoy.filters.http.router
        listenerName: inbound:127.0.0.0:8080
      value: |
        name: envoy.filters.http.gzip
  - httpFilter:
      operation: patch
      match:
        name: envoy.filters.network.tcp_proxy
        listenerName: inbound:127.0.0.0:8080
      value: |
        name: envoy.filters.http.router
        typedConfig:
          '@type': type.googleapis.com/envoy.extensions.filters.http.router.v3.Router
          dynamicStats: false
  - httpFilter:
      operation: remove
    `),
		)

		type testCase struct {
			inputYaml string
			expected  string
		}

		DescribeTable("should validate all fields and return as much individual errors as possible",
			func(given testCase) {
				// setup
				resource := NewMeshProxyPatchResource()

				// when
				err := core_model.FromYAML([]byte(given.inputYaml), &resource.Spec)
				Expect(err).ToNot(HaveOccurred())
				verr := resource.Validate()

				// then
				actual, err := yaml.Marshal(verr)
				Expect(err).ToNot(HaveOccurred())
				Expect(actual).To(MatchYAML(given.expected))
			},
			Entry("empty default", testCase{
				inputYaml: `
targetRef:
  kind: Mesh
`,
				expected: `
                violations:
                - field: spec.default.appendModifications
                  message: must not be empty`,
			}),
			Entry("empty modification", testCase{
				inputYaml: `
targetRef:
  kind: Mesh
default:
  appendModifications:
  - {}
`,
				expected: `
                violations:
                - field: spec.default.appendModifications[0]
                  message: at least one modification has to be defined`,
			}),
			Entry("invalid cluster modifications", testCase{
				inputYaml: `
targetRef:
  kind: Mesh
default:
  appendModifications:
  - cluster:
      operation: addFirst
  - cluster:
      operation: add
      value: '{'
  - cluster:
      operation: patch
      value: '{'
  - cluster:
      operation: add
      match:
        name: inbound:127.0.0.1:80
      value: |
        name: xyz
  - cluster:
      operation: remove
      value: |
        name: xyz
  - cluster:
      operation: add
  - cluster:
      operation: patch
`,
				expected: `
                violations:
                - field: spec.default.appendModifications[0].cluster.operation
                  message: 'invalid operation. Available operations: "add", "patch", "remove"'
                - field: spec.default.appendModifications[1].cluster.value
                  message: 'native Envoy resource is not valid: unexpected EOF'
                - field: spec.default.appendModifications[2].cluster.value
                  message: 'native Envoy resource is not valid: unexpected EOF'
                - field: spec.default.appendModifications[3].cluster.match
                  message: must not be defined
                - field: spec.default.appendModifications[4].cluster.value
                  message: must not be defined
                - field: spec.default.appendModifications[5].cluster.value
                  message: must be defined
                - field: spec.default.appendModifications[6].cluster.value
                  message: must be defined`,
			}),
			Entry("invalid listener modifications", testCase{
				inputYaml: `
targetRef:
  kind: Mesh
default:
  appendModifications:
  - listener:
      operation: addFirst
  - listener:
      operation: add
      value: '{'
  - listener:
      operation: patch
      value: '{'
  - listener:
      operation: add
      match:
        name: inbound:127.0.0.1:80
      value: |
        name: xyz
        address:
          socketAddress:
            address: 192.168.0.1
            portValue: 8080
  - listener:
      operation: remove
      value: |
        name: xyz
`,
				expected: `
                violations:
                - field: spec.default.appendModifications[0].listener.operation
                  message: 'invalid operation. Available operations: "add", "patch", "remove"'
                - field: spec.default.appendModifications[1].listener.value
                  message: 'native Envoy resource is not valid: unexpected EOF'
                - field: spec.default.appendModifications[2].listener.value
                  message: 'native Envoy resource is not valid: unexpected EOF'
                - field: spec.default.appendModifications[3].listener.match
                  message: must not be defined
                - field: spec.default.appendModifications[4].listener.value
                  message: must not be defined`,
			}),
			Entry("invalid network filter operation", testCase{
				inputYaml: `
targetRef:
  kind: Mesh
default:
  appendModifications:
  - networkFilter:
      operation: addFirst
      value: '{'
  - networkFilter:
      operation: addBefore
      value: '{'
  - networkFilter:
      operation: addAfter
      value: '{'
  - networkFilter:
      operation: patch
      value: '{'
  - networkFilter:
      operation: remove
      value: '{'
  - networkFilter:
      operation: add
`,
				expected: `
                violations:
                - field: spec.default.appendModifications[0].networkFilter.value
                  message: 'native Envoy resource is not valid: unexpected EOF'
                - field: spec.default.appendModifications[1].networkFilter.match.name
                  message: must be defined. You need to pick a filter before which this one will be added
                - field: spec.default.appendModifications[1].networkFilter.value
                  message: 'native Envoy resource is not valid: unexpected EOF'
                - field: spec.default.appendModifications[2].networkFilter.match.name
                  message: must be defined. You need to pick a filter after which this one will be added
                - field: spec.default.appendModifications[2].networkFilter.value
                  message: 'native Envoy resource is not valid: unexpected EOF'
                - field: spec.default.appendModifications[3].networkFilter.match.name
                  message: must be defined
                - field: spec.default.appendModifications[3].networkFilter.value
                  message: 'native Envoy resource is not valid: unexpected EOF'
                - field: spec.default.appendModifications[4].networkFilter.value
                  message: must not be defined
                - field: spec.default.appendModifications[5].networkFilter.operation
                  message: 'invalid operation. Available operations: "addFirst", "addLast", "addBefore", "addAfter", "patch", "remove"'`,
			}),
			Entry("invalid http filter operation", testCase{
				inputYaml: `
targetRef:
  kind: Mesh
default:
  appendModifications:
  - httpFilter:
      operation: addFirst
      value: '{'
  - httpFilter:
      operation: addBefore
      value: '{'
  - httpFilter:
      operation: addAfter
      value: '{'
  - httpFilter:
      operation: patch
      value: '{'
  - httpFilter:
      operation: remove
      value: '{'
  - httpFilter:
      operation: add
`,
				expected: `
                violations:
                - field: spec.default.appendModifications[0].httpFilter.value
                  message: 'native Envoy resource is not valid: unexpected EOF'
                - field: spec.default.appendModifications[1].httpFilter.match.name
                  message: must be defined. You need to pick a filter before which this one will be added
                - field: spec.default.appendModifications[1].httpFilter.value
                  message: 'native Envoy resource is not valid: unexpected EOF'
                - field: spec.default.appendModifications[2].httpFilter.match.name
                  message: must be defined. You need to pick a filter after which this one will be added
                - field: spec.default.appendModifications[2].httpFilter.value
                  message: 'native Envoy resource is not valid: unexpected EOF'
                - field: spec.default.appendModifications[3].httpFilter.match.name
                  message: must be defined
                - field: spec.default.appendModifications[3].httpFilter.value
                  message: 'native Envoy resource is not valid: unexpected EOF'
                - field: spec.default.appendModifications[4].httpFilter.value
                  message: must not be defined
                - field: spec.default.appendModifications[5].httpFilter.operation
                  message: 'invalid operation. Available operations: "addFirst", "addLast", "addBefore", "addAfter", "patch", "remove"'`,
			}),
			Entry("invalid virtual host operation", testCase{
				inputYaml: `
targetRef:
  kind: Mesh
default:
  appendModifications:
  - virtualHost:
      operation: add
      match:
        name: xyz
      value: '{'
  - virtualHost:
      operation: addFirst
  - virtualHost:
      operation: patch
      value: '{'
  - virtualHost:
      operation: remove
      value: '{'
`,
				expected: `
                violations:
                - field: spec.default.appendModifications[0].virtualHost.match.name
                  message: must not be defined
                - field: spec.default.appendModifications[0].virtualHost.value
                  message: 'native Envoy resource is not valid: unexpected EOF'
                - field: spec.default.appendModifications[1].virtualHost.operation
                  message: 'invalid operation. Available operations: "add", "patch", "remove"'
                - field: spec.default.appendModifications[2].virtualHost.value
                  message: 'native Envoy resource is not valid: unexpected EOF'
                - field: spec.default.appendModifications[3].virtualHost.value
                  message: must not be defined`,
			}),
		)
	})
})
