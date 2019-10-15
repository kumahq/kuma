package generate_test

import (
	"bytes"
	"fmt"
	"github.com/Kong/kuma/app/kumactl/cmd"
	kumactl_cmd "github.com/Kong/kuma/app/kumactl/pkg/cmd"
	"github.com/Kong/kuma/app/kumactl/pkg/tokens"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/spf13/cobra"
)

type staticDataplaneTokenGenerator struct {
}

var _ tokens.DataplaneTokenClient = &staticDataplaneTokenGenerator{}

func (s *staticDataplaneTokenGenerator) Generate(name string, mesh string) (string, error) {
	return fmt.Sprintf("token-for-%s-%s", name, mesh), nil
}

var _ = Describe("kumactl generate dataplane-token", func() {

	var rootCmd *cobra.Command
	var buf *bytes.Buffer

	BeforeEach(func() {
		ctx := kumactl_cmd.RootContext{
			Runtime: kumactl_cmd.RootRuntime{
				NewDataplaneTokenClient: func(string) (tokens.DataplaneTokenClient, error) {
					return &staticDataplaneTokenGenerator{}, nil
				},
			},
		}

		rootCmd = cmd.NewRootCmd(&ctx)
		buf = &bytes.Buffer{}
		rootCmd.SetOut(buf)
	})

	It("should generate a token", func() {
		// when
		rootCmd.SetArgs([]string{"generate", "dataplane-token", "--name=example", "--mesh=pilot"})
		err := rootCmd.Execute()

		// then
		Expect(err).ToNot(HaveOccurred())

		// and
		Expect(buf.String()).To(Equal("token-for-example-pilot"))
	})

	It("should generate a token for default mesh when it is not specified", func() {
		// when
		rootCmd.SetArgs([]string{"generate", "dataplane-token", "--name=example"})
		err := rootCmd.Execute()

		// then
		Expect(err).ToNot(HaveOccurred())

		// and
		Expect(buf.String()).To(Equal("token-for-example-default"))
	})
})
