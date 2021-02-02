package generator_test

import (
	"io/ioutil"
	"path/filepath"

	"github.com/kumahq/kuma/pkg/tls"

	"github.com/kumahq/kuma/pkg/test/runtime"

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
	// overridable by unit tests
	generator.NewSelfSignedCert = func(commonName string, certType tls.CertType, hosts ...string) (tls.KeyPair, error) {
		return tls.KeyPair{
			CertPEM: []byte("LS0=="),
			KeyPEM:  []byte("LS0=="),
		}, nil
	}

	generator := generator.AdminProxyGenerator{}

	type testCase struct {
		dataplaneFile string
		expected      string
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
				ControlPlane:     nil,
				Mesh:             context.MeshContext{},
				EnvoyAdminClient: &runtime.DummyEnvoyAdminClient{},
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

			// and output matches golden files
			ExpectMatchesGoldenFiles(actual, filepath.Join("testdata", "admin", given.expected))
		},
		Entry("should generate admin resources", testCase{
			dataplaneFile: "01.dataplane.input.yaml",
			expected:      "01.envoy-config.golden.yaml",
		}),
	)
})
