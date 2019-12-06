package ca_test

import (
	"bytes"
	"github.com/Kong/kuma/pkg/core/ca/provided/rest"
	"github.com/Kong/kuma/pkg/core/ca/provided/rest/types"
	"github.com/Kong/kuma/pkg/tls"
	"io/ioutil"
	"path/filepath"

	"github.com/Kong/kuma/app/kumactl/cmd"
	kumactl_cmd "github.com/Kong/kuma/app/kumactl/pkg/cmd"
	"github.com/spf13/cobra"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ rest.ProvidedCaClient = &staticProvidedCaClient{}

type staticProvidedCaClient struct {
	addMesh string
	addPair tls.KeyPair

	deleteCertMesh string
	deleteCertId   string

	deleteCaMesh string

	signingCerts     []types.SigningCert
	signingCertsMesh string
}

func (s *staticProvidedCaClient) AddSigningCertificate(mesh string, pair tls.KeyPair) (types.SigningCert, error) {
	s.addMesh = mesh
	s.addPair = pair
	return types.SigningCert{
		Id: "id-13456",
	}, nil
}

func (s *staticProvidedCaClient) DeleteSigningCertificate(mesh string, id string) error {
	s.deleteCertMesh = mesh
	s.deleteCertId = id
	return nil
}

func (s *staticProvidedCaClient) DeleteCa(mesh string) error {
	s.deleteCaMesh = mesh
	return nil
}

func (s *staticProvidedCaClient) SigningCertificates(mesh string) ([]types.SigningCert, error) {
	s.signingCertsMesh = mesh
	return s.signingCerts, nil
}

var _ = Describe("kumactl manage provided ca", func() {

	var rootCtx *kumactl_cmd.RootContext
	var rootCmd *cobra.Command
	var buf *bytes.Buffer
	var client *staticProvidedCaClient

	BeforeEach(func() {
		client = &staticProvidedCaClient{}
		rootCtx = &kumactl_cmd.RootContext{
			Runtime: kumactl_cmd.RootRuntime{
				NewProvidedCaClient: func(_ string) (rest.ProvidedCaClient, error) {
					return client, nil
				},
			},
		}

		rootCmd = cmd.NewRootCmd(rootCtx)
		buf = &bytes.Buffer{}
		rootCmd.SetOut(buf)
	})

	It("should add certificate", func() {
		// setup
		certBytes, err := ioutil.ReadFile(filepath.Join("testdata", "cert.pem"))
		Expect(err).ToNot(HaveOccurred())
		keyBytes, err := ioutil.ReadFile(filepath.Join("testdata", "cert.key"))
		Expect(err).ToNot(HaveOccurred())
		keyPar := tls.KeyPair{
			CertPEM: certBytes,
			KeyPEM:  keyBytes,
		}

		// given
		rootCmd.SetArgs([]string{
			"manage", "ca", "provided", "certificates", "add",
			"--mesh", "demo",
			"--key-file", filepath.Join("testdata", "cert.key"),
			"--cert-file", filepath.Join("testdata", "cert.pem"),
		})

		// when
		err = rootCmd.Execute()
		// then
		Expect(err).ToNot(HaveOccurred())

		// and
		Expect(client.addMesh).To(Equal("demo"))
		Expect(client.addPair).To(Equal(keyPar))
	})

	It("should delete certificate", func() {
		// given
		rootCmd.SetArgs([]string{
			"manage", "ca", "provided", "certificates", "delete",
			"--mesh", "demo",
			"--id", "1234-5678",
		})

		// when
		err := rootCmd.Execute()
		// then
		Expect(err).ToNot(HaveOccurred())

		// and
		Expect(client.deleteCertId).To(Equal("1234-5678"))
		Expect(client.deleteCertMesh).To(Equal("demo"))
	})

	It("should list certificates", func() {
		client.signingCerts = []types.SigningCert{
			{
				Id:   "1234",
				Cert: "ABCD",
			},
			{
				Id:   "4321",
				Cert: "ZAQWSX",
			},
		}

		// given
		rootCmd.SetArgs([]string{
			"manage", "ca", "provided", "certificates", "list",
			"--mesh", "demo",
		})

		// when
		err := rootCmd.Execute()
		// then
		Expect(err).ToNot(HaveOccurred())

		// and
		Expect(buf.String()).To(Equal(`ID     CERT
1234   ABCD
4321   ZAQWSX
`))
	})

	It("should delete certificate authority", func() {
		// given
		rootCmd.SetArgs([]string{
			"manage", "ca", "provided", "delete",
			"--mesh", "demo",
		})

		// when
		err := rootCmd.Execute()
		// then
		Expect(err).ToNot(HaveOccurred())

		// and
		Expect(client.deleteCaMesh).To(Equal("demo"))
	})
})
