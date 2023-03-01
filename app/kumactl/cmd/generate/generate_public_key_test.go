package generate_test

import (
	"bytes"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	kumactl_cmd "github.com/kumahq/kuma/app/kumactl/cmd"
	"github.com/kumahq/kuma/app/kumactl/pkg/cmd"
)

var _ = Describe("Generate Public Key", func() {
	It("should generate public key", func() {
		// setup
		ctx := cmd.DefaultRootContext()

		rootCmd := kumactl_cmd.NewRootCmd(ctx)
		buf := &bytes.Buffer{}
		rootCmd.SetOut(buf)

		// given
		rootCmd.SetArgs([]string{"generate", "public-key", "--signing-key-path", filepath.Join("..", "..", "..", "..", "test", "keys", "samplekey.pem")})

		// when
		err := rootCmd.Execute()

		// then
		Expect(err).ToNot(HaveOccurred())
		publicKeyBytes, err := os.ReadFile(filepath.Join("..", "..", "..", "..", "test", "keys", "publickey.pem"))
		Expect(err).ToNot(HaveOccurred())
		Expect(buf.String()).To(Equal(string(publicKeyBytes)))
	})
})
