package generate_test

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/app/kumactl/cmd/generate"
	"github.com/kumahq/kuma/pkg/tls"
	"github.com/kumahq/kuma/pkg/util/test"
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

	AssertOutput := func(keyPath string, certPath string) {
		Expect(stdout.String()).To(ContainSubstring(
			fmt.Sprintf("Private key saved in %s", keyFile.Name())),
		)
		Expect(stdout.String()).To(ContainSubstring(
			fmt.Sprintf("Certificate saved in %s", certFile.Name())),
		)
	}

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
			generate.NewSelfSignedCert = func(commonName string, certType tls.CertType, hosts ...string) (tls.KeyPair, error) {
				Expect(commonName).To(Equal("client-name"))
				Expect(certType).To(Equal(tls.ClientCertType))
				Expect(hosts).To(ConsistOf("client-name"))
				return tls.KeyPair{
					CertPEM: []byte("CERT"),
					KeyPEM:  []byte("KEY"),
				}, nil
			}
		})

		It("should generate client certificate", func() {
			// given
			rootCmd := test.DefaultTestingRootCmd()
			rootCmd.SetArgs([]string{"generate", "tls-certificate",
				"--key-file", keyFile.Name(),
				"--cert-file", certFile.Name(),
				"--type", "client",
				"--hostname", "client-name",
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
			AssertOutput(keyFile.Name(), certFile.Name())
		})

		It("should validate that --hostname is present", func() {
			// given
			rootCmd := test.DefaultTestingRootCmd()
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
			Expect(err).To(MatchError("required flag(s) \"hostname\" not set"))
		})

	})

	Context("server certificate", func() {
		BeforeEach(func() {
			generate.NewSelfSignedCert = func(commonName string, certType tls.CertType, hosts ...string) (tls.KeyPair, error) {
				Expect(commonName).To(Equal("kuma1.internal"))
				Expect(certType).To(Equal(tls.ServerCertType))
				Expect(hosts).To(ConsistOf("kuma1.internal", "kuma2.internal"))
				return tls.KeyPair{
					CertPEM: []byte("CERT"),
					KeyPEM:  []byte("KEY"),
				}, nil
			}
		})

		It("should generate server certificate", func() {
			// given
			rootCmd := test.DefaultTestingRootCmd()
			rootCmd.SetArgs([]string{"generate", "tls-certificate",
				"--key-file", keyFile.Name(),
				"--cert-file", certFile.Name(),
				"--type", "server",
				"--hostname", "kuma1.internal",
				"--hostname", "kuma2.internal",
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
			AssertOutput(keyFile.Name(), certFile.Name())
		})

		It("should validate that --hostname is present", func() {
			// given
			rootCmd := test.DefaultTestingRootCmd()
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
			Expect(err).To(MatchError("required flag(s) \"hostname\" not set"))
		})

	})
})
