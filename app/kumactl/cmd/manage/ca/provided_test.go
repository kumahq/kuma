package ca_test

import (
	"bytes"
	"io/ioutil"
	"path/filepath"

	"github.com/Kong/kuma/app/kumactl/cmd"
	kumactl_cmd "github.com/Kong/kuma/app/kumactl/pkg/cmd"
	"github.com/Kong/kuma/pkg/catalog"
	catalog_client "github.com/Kong/kuma/pkg/catalog/client"
	"github.com/Kong/kuma/pkg/core/ca/provided/rest"
	"github.com/Kong/kuma/pkg/core/ca/provided/rest/types"
	test_catalog "github.com/Kong/kuma/pkg/test/catalog"
	"github.com/Kong/kuma/pkg/tls"
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
		Expect(buf.String()).To(Equal(`ID     COMMON NAME   SERIAL NUMBER   NOT VALID BEFORE                NOT VALID AFTER                 SHA-1 FINGERPRINT                          SHA-256 FINGERPRINT
1234   default       0               2019-12-04 17:34:55 +0000 UTC   2029-12-01 17:35:05 +0000 UTC   e85e054b40e4c88cb45a7ae8018aaeb9f1c21be6   3082031b30820203a003020102020100300d06092a864886f70d01010b05003030310d300b060355040a13044b756d61310d300b060355040b13044d6573683110300e0603550403130764656661756c74301e170d3139313230343137333435355a170d3239313230313137333530355a3030310d300b060355040a13044b756d61310d300b060355040b13044d6573683110300e0603550403130764656661756c7430820122300d06092a864886f70d01010105000382010f003082010a02820101009d10433e4de9c501e22b8a33a0f7a1850230033fc3166f92aaf4e6a4319e6906035dc358210962ddeda8e3f52b43ee1c36ea0f30b4450514cf0a157cfacf0bc738cd2ecaf0e1a618997bfdc1ea80500efe166e58676ffdf7856ce87107e8921e3547356999db6256018a5f4a3c970e3f444d62b1519e46a09245208942f754d0f0e8a86cda0daa3dbfe0d3a3129849825307182d146600182c26daa4d68ba985ff4657cf578979fe428812991008309be73daa26caea8fbed3fab432c5e3bdb3673a866761aec7c5715ec637802a2a8aadca740509acefba3906288b3b76a11d6590d38c8b42e2dddb372e42fcb1f2ba524b4d379fe7d2d3be863c2928d5dbc90203010001a340303e300e0603551d0f0101ff040403020106300f0603551d130101ff040530030101ff301b0603551d110414301286107370696666653a2f2f64656661756c74300d06092a864886f70d01010b050003820101000b8fde1a98cfcd1b49efa736396ab79ab98b979e7a611ffa7110e105f7b4d2d85bd7d4448d051439ac6cc63fced73aa1832c03284ae41bfc6e819fc0535dbd327232ec2da3dc23dfed198d2a8d7e2072cb0ce6e5ac45f98fe86321c45ca46c837afebd4e021dd6559f687d7cd734e7f5f8a0bc5d316a97da3522b632e2b4ce173ef4471cae9f42d64245261284cb0ce35ce27842adec503ce52466e307431ca0dff51fcf9ec9062d25db3d1f580c60b738ec31badc3a6c72e8ebb66f97f0103c5d35911470fd7096f1d3a4cc7afd632f89f65f9a8b201f0261e8ce3792334f49ad034a5d4531207145025471bf20a484e5c8de24b9726e08ff10962654c2c310e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855
4321   default       0               2019-12-04 17:34:55 +0000 UTC   2029-12-01 17:35:05 +0000 UTC   e85e054b40e4c88cb45a7ae8018aaeb9f1c21be6   3082031b30820203a003020102020100300d06092a864886f70d01010b05003030310d300b060355040a13044b756d61310d300b060355040b13044d6573683110300e0603550403130764656661756c74301e170d3139313230343137333435355a170d3239313230313137333530355a3030310d300b060355040a13044b756d61310d300b060355040b13044d6573683110300e0603550403130764656661756c7430820122300d06092a864886f70d01010105000382010f003082010a02820101009d10433e4de9c501e22b8a33a0f7a1850230033fc3166f92aaf4e6a4319e6906035dc358210962ddeda8e3f52b43ee1c36ea0f30b4450514cf0a157cfacf0bc738cd2ecaf0e1a618997bfdc1ea80500efe166e58676ffdf7856ce87107e8921e3547356999db6256018a5f4a3c970e3f444d62b1519e46a09245208942f754d0f0e8a86cda0daa3dbfe0d3a3129849825307182d146600182c26daa4d68ba985ff4657cf578979fe428812991008309be73daa26caea8fbed3fab432c5e3bdb3673a866761aec7c5715ec637802a2a8aadca740509acefba3906288b3b76a11d6590d38c8b42e2dddb372e42fcb1f2ba524b4d379fe7d2d3be863c2928d5dbc90203010001a340303e300e0603551d0f0101ff040403020106300f0603551d130101ff040530030101ff301b0603551d110414301286107370696666653a2f2f64656661756c74300d06092a864886f70d01010b050003820101000b8fde1a98cfcd1b49efa736396ab79ab98b979e7a611ffa7110e105f7b4d2d85bd7d4448d051439ac6cc63fced73aa1832c03284ae41bfc6e819fc0535dbd327232ec2da3dc23dfed198d2a8d7e2072cb0ce6e5ac45f98fe86321c45ca46c837afebd4e021dd6559f687d7cd734e7f5f8a0bc5d316a97da3522b632e2b4ce173ef4471cae9f42d64245261284cb0ce35ce27842adec503ce52466e307431ca0dff51fcf9ec9062d25db3d1f580c60b738ec31badc3a6c72e8ebb66f97f0103c5d35911470fd7096f1d3a4cc7afd632f89f65f9a8b201f0261e8ce3792334f49ad034a5d4531207145025471bf20a484e5c8de24b9726e08ff10962654c2c310e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855
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
		Expect(buf.String()).To(Equal(`deleted certificate authority for mesh "demo"`))
	})
})
