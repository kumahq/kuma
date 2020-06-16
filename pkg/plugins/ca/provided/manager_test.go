package provided_test

import (
	"context"
	"crypto/x509"
	"encoding/pem"
	"io/ioutil"
	"path/filepath"
	"time"

	"github.com/ghodss/yaml"
	structpb "github.com/golang/protobuf/ptypes/struct"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	mesh_proto "github.com/Kong/kuma/api/mesh/v1alpha1"
	system_proto "github.com/Kong/kuma/api/system/v1alpha1"
	"github.com/Kong/kuma/pkg/core"
	core_ca "github.com/Kong/kuma/pkg/core/ca"
	"github.com/Kong/kuma/pkg/core/datasource"
	"github.com/Kong/kuma/pkg/plugins/ca/provided"
	provided_config "github.com/Kong/kuma/pkg/plugins/ca/provided/config"
	"github.com/Kong/kuma/pkg/util/proto"
)

var _ = Describe("Provided CA", func() {
	var caManager core_ca.Manager

	now := time.Now()

	BeforeEach(func() {
		core.Now = func() time.Time {
			return now
		}
		caManager = provided.NewProvidedCaManager(datasource.NewDataSourceLoader(nil))
	})

	AfterEach(func() {
		core.Now = time.Now
	})

	Context("ValidateBackend", func() {
		type testCase struct {
			configYAML string
			expected   string
		}

		DescribeTable("should Validate invalid config",
			func(given testCase) {
				// given
				str := structpb.Struct{}
				err := proto.FromYAML([]byte(given.configYAML), &str)
				Expect(err).ToNot(HaveOccurred())

				// when
				verr := caManager.ValidateBackend(context.Background(), "default", mesh_proto.CertificateAuthorityBackend{
					Name: "provided-1",
					Type: "provided",
					Conf: &str,
				})

				// then
				actual, err := yaml.Marshal(verr)
				Expect(err).ToNot(HaveOccurred())
				Expect(actual).To(MatchYAML(given.expected))
			},
			Entry("empty config", testCase{
				configYAML: ``,
				expected: `
            violations:
            - field: cert
              message: has to be defined
            - field: key
              message: has to be defined`,
			}),
			Entry("config without data source", testCase{
				configYAML: `
            cert: {}
            key: {}`,
				expected: `
            violations:
            - field: cert
              message: 'data source has to be chosen. Available sources: secret, file, inline'
            - field: key
              message: 'data source has to be chosen. Available sources: secret, file, inline'`,
			}),
			Entry("config with empty secret", testCase{
				configYAML: `
            cert:
              secret:
            key:
              secret:`,
				expected: `
            violations:
            - field: cert.secret
              message: cannot be empty
            - field: key.secret
              message: cannot be empty`,
			}),
			Entry("config with empty secret", testCase{
				configYAML: `
            cert:
              file: '/tmp/non-existing-file'
            key:
              file: /tmp/non-existing-file`,
				expected: `
            violations:
            - field: cert
              message: 'could not load data: open /tmp/non-existing-file: no such file or directory'
            - field: key
              message: 'could not load data: open /tmp/non-existing-file: no such file or directory'`,
			}),
			Entry("config with invalid cert", testCase{
				configYAML: `
            cert:
              inline: dGVzdA==
            key:
              inline: dGVzdA==`,
				expected: `
            violations:
            - field: cert
              message: 'not a valid TLS key pair: tls: failed to find any PEM data in certificate input'`,
			}),
		)
	})

	var backendWithTestCerts mesh_proto.CertificateAuthorityBackend
	var backendWithInvalidCerts mesh_proto.CertificateAuthorityBackend

	BeforeEach(func() {
		cfg := provided_config.ProvidedCertificateAuthorityConfig{
			Cert: &system_proto.DataSource{
				Type: &system_proto.DataSource_File{
					File: filepath.Join("testdata", "ca.pem"),
				},
			},
			Key: &system_proto.DataSource{
				Type: &system_proto.DataSource_File{
					File: filepath.Join("testdata", "ca.key"),
				},
			},
		}
		str, err := proto.ToStruct(&cfg)
		Expect(err).ToNot(HaveOccurred())

		backendWithTestCerts = mesh_proto.CertificateAuthorityBackend{
			Name: "provided-1",
			Type: "provided",
			Conf: &str,
			DpCert: &mesh_proto.CertificateAuthorityBackend_DpCert{
				Rotation: &mesh_proto.CertificateAuthorityBackend_DpCert_Rotation{
					Expiration: "1s",
				},
			},
		}

		invalidCfg := provided_config.ProvidedCertificateAuthorityConfig{
			Cert: &system_proto.DataSource{
				Type: &system_proto.DataSource_File{
					File: filepath.Join("testdata", "invalid.pem"),
				},
			},
			Key: &system_proto.DataSource{
				Type: &system_proto.DataSource_File{
					File: filepath.Join("testdata", "invalid.key"),
				},
			},
		}
		invalidStr, err := proto.ToStruct(&invalidCfg)
		Expect(err).ToNot(HaveOccurred())

		backendWithInvalidCerts = mesh_proto.CertificateAuthorityBackend{
			Name: "provided-2",
			Type: "provided",
			Conf: &invalidStr,
		}
	})

	Context("GetRootCert", func() {
		It("should load return root certs", func() {
			// given
			expectedCert, err := ioutil.ReadFile(filepath.Join("testdata", "ca.pem"))
			Expect(err).ToNot(HaveOccurred())

			// when
			rootCerts, err := caManager.GetRootCert(context.Background(), "default", backendWithTestCerts)

			// then
			Expect(err).ToNot(HaveOccurred())
			Expect(rootCerts).To(HaveLen(1))
			Expect(rootCerts[0]).To(Equal(expectedCert))
		})

		It("should throw an error on invalid certs", func() {
			// when
			_, err := caManager.GetRootCert(context.Background(), "default", backendWithInvalidCerts)

			// then
			Expect(err).To(MatchError(`failed to load CA key pair for Mesh "default" and backend "provided-2": could not load data: open testdata/invalid.key: no such file or directory`))
		})
	})

	Context("GenerateDataplaneCert", func() {
		It("should generate dataplane cert", func() {
			// when
			pair, err := caManager.GenerateDataplaneCert(context.Background(), "default", backendWithTestCerts, []string{"web", "web-api"})

			// then
			Expect(err).ToNot(HaveOccurred())
			Expect(pair.KeyPEM).ToNot(BeEmpty())
			Expect(pair.CertPEM).ToNot(BeEmpty())

			// and should generate cert for dataplane with spiffe URI
			block, _ := pem.Decode(pair.CertPEM)
			cert, err := x509.ParseCertificate(block.Bytes)
			Expect(err).ToNot(HaveOccurred())
			Expect(cert.URIs).To(HaveLen(2))
			Expect(cert.URIs[0].String()).To(Equal("spiffe://default/web"))
			Expect(cert.URIs[1].String()).To(Equal("spiffe://default/web-api"))
			Expect(cert.NotAfter).To(Equal(now.UTC().Truncate(time.Second).Add(1 * time.Second))) // time in cert is in UTC and truncated to seconds
		})

		It("should throw an error on invalid certs", func() {
			// when
			_, err := caManager.GenerateDataplaneCert(context.Background(), "default", backendWithInvalidCerts, []string{"web"})

			// then
			Expect(err).To(MatchError(`failed to load CA key pair for Mesh "default" and backend "provided-2": could not load data: open testdata/invalid.key: no such file or directory`))
		})
	})

	Context("UsedSecret", func() {
		It("should return empty list when no secrets are used", func() {
			// when
			secrets, err := caManager.UsedSecrets("default", backendWithTestCerts)

			// then
			Expect(err).ToNot(HaveOccurred())
			Expect(secrets).To(BeEmpty())
		})

		It("should return list of secrets", func() {
			// given
			backend := mesh_proto.CertificateAuthorityBackend{
				Name: "provided-1",
				Type: "provided",
				Conf: proto.MustToStruct(&provided_config.ProvidedCertificateAuthorityConfig{
					Cert: &system_proto.DataSource{
						Type: &system_proto.DataSource_Secret{
							Secret: "cert-sec",
						},
					},
					Key: &system_proto.DataSource{
						Type: &system_proto.DataSource_Secret{
							Secret: "key-sec",
						},
					},
				}),
			}

			// when
			secrets, err := caManager.UsedSecrets("default", backend)

			// then
			Expect(err).ToNot(HaveOccurred())
			Expect(secrets).To(HaveLen(2))
			Expect(secrets[0]).To(Equal("cert-sec"))
			Expect(secrets[1]).To(Equal("key-sec"))
		})
	})
})
