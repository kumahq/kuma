package generate_test

import (
	"bytes"
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
		rootCmd.SetArgs([]string{"generate", "public-key", "--signing-key-path", filepath.Join("testdata", "samplekey.pem")})

		// when
		err := rootCmd.Execute()

		// then
		Expect(err).ToNot(HaveOccurred())
		Expect(buf.String()).To(Equal(`-----BEGIN RSA PUBLIC KEY-----
MIIBCgKCAQEAqwbFZ7LSuRGEkFPsZOLYuimsjDeie4sdtqIVW9bLDrTSql+o2sBL
wt22MJ897/oq7+jZhVlENE1ddAKdFSWv3nhOI/XK9VJt7qNudcoC9252XrycIi5h
i700CDgdRgRt+2paZiRCgc5afNMHJmVIp2d2lQTUKn/pQGlqY4ufuA3U1z+8t++k
oGnj0sKIcXzqa5ZZxZ/81khp0e0Ze7llTmfEU3gQXu/Coa2y7LEUHdrNalM3si0v
FvX0KmBtADEJ4n9Jo4ja3hDmp83Q4KjJq0xKbhh9Fp3AjwjDb0fVFwbt+8SdVgyV
5PE+7HdigwlJ/cOVb9IY/UKVgCzlW5inCQIDAQAB
-----END RSA PUBLIC KEY-----
`))
	})
})
