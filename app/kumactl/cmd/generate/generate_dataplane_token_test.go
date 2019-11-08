package generate_test

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/Kong/kuma/app/kumactl/cmd"
	kumactl_cmd "github.com/Kong/kuma/app/kumactl/pkg/cmd"
	"github.com/Kong/kuma/app/kumactl/pkg/tokens"
	"github.com/Kong/kuma/pkg/catalogue"
	catalogue_client "github.com/Kong/kuma/pkg/catalogue/client"
	config_kumactl "github.com/Kong/kuma/pkg/config/app/kumactl/v1alpha1"
	test_catalogue "github.com/Kong/kuma/pkg/test/catalogue"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/spf13/cobra"
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
				NewDataplaneTokenClient: func(string, *config_kumactl.Context_DataplaneTokenApiCredentials) (tokens.DataplaneTokenClient, error) {
					return generator, nil
				},
				NewCatalogueClient: func(s string) (catalogue_client.CatalogueClient, error) {
					return &test_catalogue.StaticCatalogueClient{
						Resp: catalogue.Catalogue{
							Apis: catalogue.Apis{
								DataplaneToken: catalogue.DataplaneTokenApi{
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
		rootCmd.SetArgs([]string{"generate", "dataplane-token", "--dataplane=example", "--mesh=pilot"})
		err := rootCmd.Execute()

		// then
		Expect(err).ToNot(HaveOccurred())

		// and
		Expect(buf.String()).To(Equal("token-for-example-pilot"))
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
		ctx.Runtime.NewCatalogueClient = func(s string) (catalogue_client.CatalogueClient, error) {
			return &test_catalogue.StaticCatalogueClient{
				Resp: catalogue.Catalogue{
					Apis: catalogue.Apis{
						DataplaneToken: catalogue.DataplaneTokenApi{
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
		Expect(buf.String()).To(Equal("Error: failed to create dataplane token client: Dataplane Token Server in Kuma CP is disabled. Either enable the server or run Kuma DP without token.\n"))
	})
})
