package generate_test

import (
	"bytes"
	"errors"
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	"github.com/spf13/cobra"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/app/kumactl/cmd"
	kumactl_cmd "github.com/kumahq/kuma/app/kumactl/pkg/cmd"
	"github.com/kumahq/kuma/app/kumactl/pkg/tokens"
	"github.com/kumahq/kuma/pkg/catalog"
	catalog_client "github.com/kumahq/kuma/pkg/catalog/client"
	config_proto "github.com/kumahq/kuma/pkg/config/app/kumactl/v1alpha1"
	test_catalog "github.com/kumahq/kuma/pkg/test/catalog"
)

type staticDataplaneTokenGenerator struct {
	err error
}

var _ tokens.DataplaneTokenClient = &staticDataplaneTokenGenerator{}

func (s *staticDataplaneTokenGenerator) Generate(name string, mesh string, tags map[string][]string, dpType string) (string, error) {
	if s.err != nil {
		return "", s.err
	}
	return fmt.Sprintf("token-for-%s-%s-%s-%s", name, mesh, mesh_proto.MultiValueTagSetFrom(tags).String(), dpType), nil
}

var _ = Describe("kumactl generate dataplane-token", func() {

	var rootCmd *cobra.Command
	var buf *bytes.Buffer
	var generator *staticDataplaneTokenGenerator
	var ctx *kumactl_cmd.RootContext

	BeforeEach(func() {
		generator = &staticDataplaneTokenGenerator{}
		ctx = &kumactl_cmd.RootContext{
			Runtime: kumactl_cmd.RootRuntime{
				NewDataplaneTokenClient: func(*config_proto.ControlPlaneCoordinates_ApiServer) (tokens.DataplaneTokenClient, error) {
					return generator, nil
				},
				NewCatalogClient: func(s string) (catalog_client.CatalogClient, error) {
					return &test_catalog.StaticCatalogClient{
						Resp: catalog.Catalog{
							Apis: catalog.Apis{
								DataplaneToken: catalog.DataplaneTokenApi{
									LocalUrl: "http://localhost:1234",
								},
							},
						},
					}, nil
				},
			},
		}

		rootCmd = cmd.NewRootCmd(ctx)
		buf = &bytes.Buffer{}
		rootCmd.SetOut(buf)
	})

	type testCase struct {
		args   []string
		result string
	}
	DescribeTable("should generate token",
		func(given testCase) {
			// when
			rootCmd.SetArgs(given.args)
			err := rootCmd.Execute()

			// then
			Expect(err).ToNot(HaveOccurred())

			// and
			Expect(buf.String()).To(Equal(given.result))
		},
		Entry("for default mesh when it is not specified", testCase{
			args:   []string{"generate", "dataplane-token", "--dataplane=example"},
			result: "token-for-example-default--",
		}),
		Entry("for all arguments", testCase{
			args:   []string{"generate", "dataplane-token", "--mesh=demo", "--dataplane=example", "--type=dataplane", "--tag", "kuma.io/service=web"},
			result: "token-for-example-demo-kuma.io/service=web-dataplane",
		}),
	)

	It("should write error when generating token fails", func() {
		// setup
		generator.err = errors.New("could not connect to API")

		// when
		rootCmd.SetArgs([]string{"generate", "dataplane-token", "--dataplane=example"})
		err := rootCmd.Execute()

		// then
		Expect(err).To(HaveOccurred())

		// and
		Expect(buf.String()).To(Equal("Error: failed to generate a dataplane token: could not connect to API\n"))
	})

})
