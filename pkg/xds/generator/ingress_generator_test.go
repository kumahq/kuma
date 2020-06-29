package generator_test

import (
	"io/ioutil"
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	mesh_proto "github.com/Kong/kuma/api/mesh/v1alpha1"
	mesh_core "github.com/Kong/kuma/pkg/core/resources/apis/mesh"
	core_xds "github.com/Kong/kuma/pkg/core/xds"
	test_model "github.com/Kong/kuma/pkg/test/resources/model"
	util_proto "github.com/Kong/kuma/pkg/util/proto"
	xds_context "github.com/Kong/kuma/pkg/xds/context"
	"github.com/Kong/kuma/pkg/xds/generator"
)

var _ = Describe("IngressGenerator", func() {
	type testCase struct {
		dataplane       string
		expected        string
		outboundTargets core_xds.EndpointMap
	}

	DescribeTable("should generate Envoy xDS resources",
		func(given testCase) {
			gen := generator.IngressGenerator{}

			dataplane := mesh_proto.Dataplane{}
			Expect(util_proto.FromYAML([]byte(given.dataplane), &dataplane)).To(Succeed())

			proxy := &core_xds.Proxy{
				Id: core_xds.ProxyId{Name: "ingress", Mesh: "default"},
				Dataplane: &mesh_core.DataplaneResource{
					Meta: &test_model.ResourceMeta{
						Version: "1",
					},
					Spec: dataplane,
				},
				OutboundTargets: given.outboundTargets,
			}

			// when
			rs, err := gen.Generate(xds_context.Context{}, proxy)
			// then
			Expect(err).ToNot(HaveOccurred())

			// when
			resp, err := core_xds.ResourceList(rs).ToDeltaDiscoveryResponse()
			// then
			Expect(err).ToNot(HaveOccurred())
			// when
			actual, err := util_proto.ToYAML(resp)
			// then
			Expect(err).ToNot(HaveOccurred())

			expected, err := ioutil.ReadFile(filepath.Join("testdata", "ingress", given.expected))
			Expect(err).ToNot(HaveOccurred())
			Expect(actual).To(MatchYAML(expected))
		},
		Entry("01. basic", testCase{
			dataplane: `
            networking:
              address: 10.0.0.1
              ingress:
                availableServices:
                  - tags:
                      service: backend
                      version: v1
                      region: eu
                  - tags:
                      service: backend
                      version: v2
                      region: us
              inbound:
                - port: 10001
`,
			expected: "01.envoy.golden.yaml",
			outboundTargets: map[core_xds.ServiceName][]core_xds.Endpoint{
				"backend": {
					{
						Target: "192.168.0.1",
						Port:   2521,
						Tags: map[string]string{
							"service": "backend",
							"version": "v1",
							"region":  "eu",
						},
						Weight: 1,
					},
					{
						Target: "192.168.0.2",
						Port:   2521,
						Tags: map[string]string{
							"service": "backend",
							"version": "v2",
							"region":  "us",
						},
						Weight: 1,
					},
				},
			},
		}),
		Entry("02. empty ingress", testCase{
			dataplane: `
            networking:
              address: 10.0.0.1
              ingress: {}
              inbound:
                - port: 10001
`,
			expected:        "02.envoy.golden.yaml",
			outboundTargets: map[core_xds.ServiceName][]core_xds.Endpoint{},
		}),
	)
})
