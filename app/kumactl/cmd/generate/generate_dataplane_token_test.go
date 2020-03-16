package generate_test

import (
	"bytes"
	"errors"
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/spf13/cobra"

	"github.com/Kong/kuma/app/kumactl/cmd"
	kumactl_cmd "github.com/Kong/kuma/app/kumactl/pkg/cmd"
	"github.com/Kong/kuma/app/kumactl/pkg/tokens"
	"github.com/Kong/kuma/pkg/catalog"
	catalog_client "github.com/Kong/kuma/pkg/catalog/client"
	config_kumactl "github.com/Kong/kuma/pkg/config/app/kumactl/v1alpha1"
	test_catalog "github.com/Kong/kuma/pkg/test/catalog"
)

type staticDataplaneTokenGenerator struct {
	err error
}

var _ tokens.DataplaneTokenClient = &staticDataplaneTokenGenerator{}

func (s *staticDataplaneTokenGenerator) Generate(name string, mesh string) (string, error) {
	if s.err != nil {
		return "", s.err
	}
	return fmt.Sprintf("token-for-%s-%s", name, mesh), nil
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
				NewDataplaneTokenClient: func(string, *config_kumactl.Context_AdminApiCredentials) (tokens.DataplaneTokenClient, error) {
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

	It("should generate a token", func() {
		// when
		rootCmd.SetArgs([]string{"generate", "dataplane-token", "--dataplane=example", "--mesh=demo"})
		err := rootCmd.Execute()

		// then
		Expect(err).ToNot(HaveOccurred())

		// and
		Expect(buf.String()).To(Equal("token-for-example-demo"))
	})

	It("should generate a token for default mesh when it is not specified", func() {
		// when
		rootCmd.SetArgs([]string{"generate", "dataplane-token", "--dataplane=example"})
		err := rootCmd.Execute()

		// then
		Expect(err).ToNot(HaveOccurred())

		// and
		Expect(buf.String()).To(Equal("token-for-example-default"))
	})

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

	It("should throw an error when dataplane token server is disabled", func() {
		// setup
		ctx.Runtime.NewCatalogClient = func(s string) (catalog_client.CatalogClient, error) {
			return &test_catalog.StaticCatalogClient{
				Resp: catalog.Catalog{
					Apis: catalog.Apis{
						DataplaneToken: catalog.DataplaneTokenApi{
							LocalUrl: "", // disabled dataplane token server
						},
					},
				},
			}, nil
		}

		// when
		rootCmd.SetArgs([]string{"generate", "dataplane-token", "--dataplane=example"})
		err := rootCmd.Execute()

		// then
		Expect(err).To(HaveOccurred())

		// and
		Expect(buf.String()).To(Equal("Error: failed to create dataplane token client: Enable the server to be able to generate tokens.\n"))
	})
})
