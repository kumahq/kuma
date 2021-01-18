package generator_test

import (
	"io/ioutil"
	"path/filepath"

	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/xds"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
	"github.com/kumahq/kuma/pkg/xds/context"
	envoy_common "github.com/kumahq/kuma/pkg/xds/envoy"
	"github.com/kumahq/kuma/pkg/xds/generator"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

var _ = Describe("AdminProxyGenerator", func() {
	generator := generator.AdminProxyGenerator{}

	type testCase struct {
		dataplaneFile   string
		envoyConfigFile string
	}

	DescribeTable("should generate envoy config",
		func(given testCase) {
			// given

			// dataplane
			dataplane := core_mesh.NewDataplaneResource()
			bytes, err := ioutil.ReadFile(filepath.Join("testdata", "admin", given.dataplaneFile))
			Expect(err).ToNot(HaveOccurred())
			parseResource(bytes, dataplane)

			ctx := context.Context{
				ControlPlane: nil,
				Mesh:         context.MeshContext{},
			}

			proxy := &xds.Proxy{
				Metadata: &xds.DataplaneMetadata{
					AdminPort: 9901,
				},
				Dataplane:  dataplane,
				APIVersion: envoy_common.APIV2,
			}

			// when
			resources, err := generator.Generate(ctx, proxy)

			// then
			Expect(err).ToNot(HaveOccurred())

			resp, err := resources.List().ToDeltaDiscoveryResponse()
			Expect(err).ToNot(HaveOccurred())
			actual, err := util_proto.ToYAML(resp)
			Expect(err).ToNot(HaveOccurred())
			expected, err := ioutil.ReadFile(filepath.Join("testdata", "admin", given.envoyConfigFile))
			Expect(err).ToNot(HaveOccurred())

			Expect(actual).To(MatchYAML(expected))
		},
		Entry("should generate admin resources", testCase{
			dataplaneFile:   "01.dataplane.input.yaml",
			envoyConfigFile: "01.envoy-config.golden.yaml",
		}),
	)
})
