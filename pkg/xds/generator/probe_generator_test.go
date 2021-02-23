package generator_test

import (
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	mesh_core "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	. "github.com/kumahq/kuma/pkg/test/matchers"
	test_model "github.com/kumahq/kuma/pkg/test/resources/model"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
	envoy_common "github.com/kumahq/kuma/pkg/xds/envoy"
	"github.com/kumahq/kuma/pkg/xds/generator"
)

var _ = Describe("ProbeGenerator", func() {
	type testCase struct {
		dataplane string
		expected  string
	}

	DescribeTable("should generate Envoy xDS resources",
		func(given testCase) {
			gen := generator.ProbeProxyGenerator{}

			dataplane := &mesh_proto.Dataplane{}
			Expect(util_proto.FromYAML([]byte(given.dataplane), dataplane)).To(Succeed())

			proxy := &core_xds.Proxy{
				Dataplane: &mesh_core.DataplaneResource{
					Meta: &test_model.ResourceMeta{
						Version: "1",
					},
					Spec: dataplane,
				},
				APIVersion: envoy_common.APIV3,
			}

			// when
			rs, err := gen.Generate(xds_context.Context{}, proxy)
			// then
			Expect(err).ToNot(HaveOccurred())

			// when
			resp, err := rs.List().ToDeltaDiscoveryResponse()
			// then
			Expect(err).ToNot(HaveOccurred())
			// when
			actual, err := util_proto.ToYAML(resp)
			// then
			Expect(err).ToNot(HaveOccurred())

			// and output matches golden files
			Expect(actual).To(MatchGoldenYAML(filepath.Join("testdata", "probe", given.expected)))
		},
		Entry("base probes", testCase{
			dataplane: `
            networking:
              inbound:
              - port: 8080
            probes:
              port: 9000
              endpoints:
              - inboundPort: 8080
                inboundPath: /healthz/probe
                path: /8080/healthz/probe
`,
			expected: "01.envoy.golden.yaml",
		}),
		Entry("empty probes", testCase{
			dataplane: ``,
			expected:  "02.envoy.golden.yaml",
		}),
		Entry("no inbound for probe", testCase{
			dataplane: `
            networking:
              inbound:
              - port: 1010
            probes:
              port: 9000
              endpoints:
              - inboundPort: 8080
                inboundPath: /healthz/probe
                path: /8080/healthz/probe
`,
			expected: "03.envoy.golden.yaml",
		}),
	)
})
