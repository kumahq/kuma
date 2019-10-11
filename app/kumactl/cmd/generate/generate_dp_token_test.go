package generate_test

import (
	"bytes"
	"fmt"
	"github.com/Kong/kuma/app/kumactl/cmd"
	kumactl_cmd "github.com/Kong/kuma/app/kumactl/pkg/cmd"
	"github.com/Kong/kuma/app/kumactl/pkg/tokens"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

type staticDpTokenGenerator struct {
}

var _ tokens.DpTokenClient = &staticDpTokenGenerator{}

func (s *staticDpTokenGenerator) Generate(name string, mesh string) (string, error) {
	return fmt.Sprintf("token-for-%s-%s", name, mesh), nil
}

var _ = Describe("kumactl generate dp-token", func() {

	It("should generate a token", func() {
		// setup
		ctx := kumactl_cmd.RootContext{
			Runtime: kumactl_cmd.RootRuntime{
				NewDpTokenClient: func(string) (tokens.DpTokenClient, error) {
					return &staticDpTokenGenerator{}, nil
				},
			},
		}

		rootCmd := cmd.NewRootCmd(&ctx)
		buf := &bytes.Buffer{}
		rootCmd.SetOut(buf)

		// when
		rootCmd.SetArgs([]string{"generate", "dp-token", "--name=example", "--mesh=default"})
		err := rootCmd.Execute()

		// then
		Expect(err).ToNot(HaveOccurred())

		// and
		Expect(buf.String()).To(Equal("token-for-example-default"))
	})
})
