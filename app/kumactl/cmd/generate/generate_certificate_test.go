package generate_test

import (
	"bytes"
	"fmt"
	"github.com/Kong/kuma/app/kumactl/cmd"
	"github.com/Kong/kuma/app/kumactl/cmd/generate"
	"github.com/Kong/kuma/pkg/tls"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"io/ioutil"
)

var _ = Describe("kumactl generate certificate", func() {

	var backupNewSelfSignedCert func(string, tls.CertType, ...string) (tls.KeyPair, error)
	BeforeEach(func() {
		backupNewSelfSignedCert = generate.NewSelfSignedCert
	})
	AfterEach(func() {
		generate.NewSelfSignedCert = backupNewSelfSignedCert
	})

	BeforeEach(func() {
		generate.NewSelfSignedCert = func(string, tls.CertType, ...string) (tls.KeyPair, error) {
			return tls.KeyPair{
				CertPEM: []byte("CERT"),
				KeyPEM:  []byte("KEY"),
			}, nil
		}
	})

	It("should generate certificate", func() {
		// given
		keyFile, err := ioutil.TempFile("", "")
		Expect(err).ToNot(HaveOccurred())
		certFile, err := ioutil.TempFile("", "")
		Expect(err).ToNot(HaveOccurred())

		stdout := &bytes.Buffer{}
		stderr := &bytes.Buffer{}

		rootCmd := cmd.DefaultRootCmd()
		rootCmd.SetArgs([]string{"generate", "certificate",
			"--key", keyFile.Name(),
			"--cert", certFile.Name(),
			"--type", "client",
		})
		rootCmd.SetOut(stdout)
		rootCmd.SetErr(stderr)

		// when
		err = rootCmd.Execute()

		// then
		Expect(err).ToNot(HaveOccurred())

		// and
		keyBytes, err := ioutil.ReadAll(keyFile)
		Expect(err).ToNot(HaveOccurred())
		Expect(string(keyBytes)).To(Equal("KEY"))

		// and
		certBytes, err := ioutil.ReadAll(certFile)
		Expect(err).ToNot(HaveOccurred())
		Expect(string(certBytes)).To(Equal("CERT"))

		// and
		Expect(stdout.String()).To(Equal(fmt.Sprintf(`Certificates generated
Key path: %s
Cert path: %s
`, keyFile.Name(), certFile.Name())))
	})
})
