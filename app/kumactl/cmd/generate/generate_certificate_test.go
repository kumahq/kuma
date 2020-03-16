package generate_test

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/Kong/kuma/app/kumactl/cmd"
	"github.com/Kong/kuma/app/kumactl/cmd/generate"
	"github.com/Kong/kuma/pkg/tls"
)

var _ = Describe("kumactl generate tls-certificate", func() {

	var backupNewSelfSignedCert func(string, tls.CertType, ...string) (tls.KeyPair, error)
	BeforeEach(func() {
		backupNewSelfSignedCert = generate.NewSelfSignedCert
	})
	AfterEach(func() {
		generate.NewSelfSignedCert = backupNewSelfSignedCert
	})

	var keyFile *os.File
	var certFile *os.File

	var stdout *bytes.Buffer
	var stderr *bytes.Buffer

	BeforeEach(func() {
		key, err := ioutil.TempFile("", "")
		Expect(err).ToNot(HaveOccurred())
		keyFile = key

		cert, err := ioutil.TempFile("", "")
		Expect(err).ToNot(HaveOccurred())
		certFile = cert

		stdout = &bytes.Buffer{}
		stderr = &bytes.Buffer{}
	})

	Context("client certificate", func() {
		BeforeEach(func() {
			generate.NewSelfSignedCert = func(commonName string, certType tls.CertType, _ ...string) (tls.KeyPair, error) {
				Expect(commonName).To(Equal("kuma"))
				Expect(certType).To(Equal(tls.ClientCertType))
				return tls.KeyPair{
					CertPEM: []byte("CERT"),
					KeyPEM:  []byte("KEY"),
				}, nil
			}
		})

		It("should generate client certificate", func() {
			// given
			rootCmd := cmd.DefaultRootCmd()
			rootCmd.SetArgs([]string{"generate", "tls-certificate",
				"--key-file", keyFile.Name(),
				"--cert-file", certFile.Name(),
				"--type", "client",
			})
			rootCmd.SetOut(stdout)
			rootCmd.SetErr(stderr)

			// when
			err := rootCmd.Execute()

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
Key was saved in: %s
Cert was saved in: %s
`, keyFile.Name(), certFile.Name())))
		})

		It("should not allow to specify control plane hostname", func() {
			// given
			rootCmd := cmd.DefaultRootCmd()
			rootCmd.SetArgs([]string{"generate", "tls-certificate",
				"--key-file", keyFile.Name(),
				"--cert-file", certFile.Name(),
				"--type", "client",
				"--cp-hostname", "kuma1.internal",
			})
			rootCmd.SetOut(stdout)
			rootCmd.SetErr(stderr)

			// when
			err := rootCmd.Execute()

			// then
			Expect(err).To(MatchError(`--cp-hostname cannot be used with "client" type`))
		})
	})

	Context("server certificate", func() {
		BeforeEach(func() {
			generate.NewSelfSignedCert = func(commonName string, certType tls.CertType, hosts ...string) (tls.KeyPair, error) {
				Expect(commonName).To(Equal("kuma"))
				Expect(certType).To(Equal(tls.ServerCertType))
				Expect(hosts).To(HaveLen(3))
				Expect(hosts).To(ConsistOf("kuma1.internal", "kuma2.internal", "localhost"))
				return tls.KeyPair{
					CertPEM: []byte("CERT"),
					KeyPEM:  []byte("KEY"),
				}, nil
			}
		})

		It("should generate server certificate", func() {
			// given
			rootCmd := cmd.DefaultRootCmd()
			rootCmd.SetArgs([]string{"generate", "tls-certificate",
				"--key-file", keyFile.Name(),
				"--cert-file", certFile.Name(),
				"--type", "server",
				"--cp-hostname", "kuma1.internal",
				"--cp-hostname", "kuma2.internal",
			})
			rootCmd.SetOut(stdout)
			rootCmd.SetErr(stderr)

			// when
			err := rootCmd.Execute()

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
Key was saved in: %s
Cert was saved in: %s
`, keyFile.Name(), certFile.Name())))
		})

		It("should validate that cp-hostname is present", func() {
			// given
			rootCmd := cmd.DefaultRootCmd()
			rootCmd.SetArgs([]string{"generate", "tls-certificate",
				"--key-file", keyFile.Name(),
				"--cert-file", certFile.Name(),
				"--type", "server",
			})
			rootCmd.SetOut(stdout)
			rootCmd.SetErr(stderr)

			// when
			err := rootCmd.Execute()

			// then
			Expect(err).To(MatchError(`--cp-hostname has to be specified with "server" type`))
		})
	})
})
