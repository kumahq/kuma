package generate_test

import (
	"bytes"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	kumactl_cmd "github.com/kumahq/kuma/app/kumactl/cmd"
	"github.com/kumahq/kuma/app/kumactl/pkg/cmd"
)

var _ = Describe("Generate Signing Key", func() {

	It("should generate signing key", func() {
		// setup
		ctx := cmd.DefaultRootContext()
		ctx.GenerateContext.NewSigningKey = func() ([]byte, error) {
			return []byte("TEST"), nil
		}

		rootCmd := kumactl_cmd.NewRootCmd(ctx)
		buf := &bytes.Buffer{}
		rootCmd.SetOut(buf)

		// given
		rootCmd.SetArgs([]string{"generate", "signing-key"})

		// when
		err := rootCmd.Execute()

		// then
		Expect(err).ToNot(HaveOccurred())
		Expect(buf.String()).To(Equal("VEVTVA=="))
	})
})
