package ca_test

import (
	"bytes"
	"github.com/Kong/kuma/app/kumactl/pkg/ca"
	"io/ioutil"
	"path/filepath"

	"github.com/Kong/kuma/app/kumactl/cmd"
	kumactl_cmd "github.com/Kong/kuma/app/kumactl/pkg/cmd"
	"github.com/Kong/kuma/pkg/catalog"
	catalog_client "github.com/Kong/kuma/pkg/catalog/client"
	kumactl_config "github.com/Kong/kuma/pkg/config/app/kumactl/v1alpha1"
	"github.com/Kong/kuma/pkg/core/ca/provided/rest/types"
	test_catalog "github.com/Kong/kuma/pkg/test/catalog"
	"github.com/Kong/kuma/pkg/tls"
	"github.com/spf13/cobra"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ ca.ProvidedCaClient = &staticProvidedCaClient{}

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
				NewProvidedCaClient: func(_ string, _ *kumactl_config.Context_AdminApiCredentials) (ca.ProvidedCaClient, error) {
					return client, nil
				},
				NewCatalogClient: func(s string) (catalog_client.CatalogClient, error) {
					return &test_catalog.StaticCatalogClient{
						Resp: catalog.Catalog{
							Apis: catalog.Apis{
								Admin: catalog.AdminApi{
									LocalUrl: "http://localhost:1234",
								},
							},
						},
					}, nil
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
		Expect(buf.String()).To(Equal(`added certificate "id-13456"`))
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
		Expect(buf.String()).To(Equal(`removed certificate "1234-5678"`))
	})

	It("should list certificates", func() {
		// given
		certBytes, err := ioutil.ReadFile(filepath.Join("testdata", "cert.pem"))
		Expect(err).ToNot(HaveOccurred())
		client.signingCerts = []types.SigningCert{
			{
				Id:   "1234",
				Cert: string(certBytes),
			},
			{
				Id:   "4321",
				Cert: string(certBytes),
			},
		}

		// and
		rootCmd.SetArgs([]string{
			"manage", "ca", "provided", "certificates", "list",
			"--mesh", "demo",
		})

		// when
		err = rootCmd.Execute()
		// then
		Expect(err).ToNot(HaveOccurred())

		// and
		Expect(buf.String()).To(Equal(`ID     COMMON NAME   SERIAL NUMBER   NOT VALID BEFORE                NOT VALID AFTER                 SHA-1 FINGERPRINT
1234   default       0               2019-12-04 17:34:55 +0000 UTC   2029-12-01 17:35:05 +0000 UTC   e85e054b40e4c88cb45a7ae8018aaeb9f1c21be6
4321   default       0               2019-12-04 17:34:55 +0000 UTC   2029-12-01 17:35:05 +0000 UTC   e85e054b40e4c88cb45a7ae8018aaeb9f1c21be6
`))
	})
})
