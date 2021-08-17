package builtin_test

import (
	"context"
	"crypto/x509"
	"encoding/pem"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core"
	core_ca "github.com/kumahq/kuma/pkg/core/ca"
	"github.com/kumahq/kuma/pkg/core/resources/apis/system"
	"github.com/kumahq/kuma/pkg/core/resources/manager"
	core_store "github.com/kumahq/kuma/pkg/core/resources/store"
	"github.com/kumahq/kuma/pkg/core/secrets/cipher"
	secret_manager "github.com/kumahq/kuma/pkg/core/secrets/manager"
	"github.com/kumahq/kuma/pkg/core/secrets/store"
	"github.com/kumahq/kuma/pkg/plugins/ca/builtin"
	"github.com/kumahq/kuma/pkg/plugins/ca/builtin/config"
	"github.com/kumahq/kuma/pkg/plugins/resources/memory"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
)

var _ = Describe("Builtin CA Manager", func() {

	var secretManager manager.ResourceManager
	var caManager core_ca.Manager

	now := time.Now()

	BeforeEach(func() {
		core.Now = func() time.Time {
			return now
		}
		secretManager = secret_manager.NewSecretManager(store.NewSecretStore(memory.NewStore()), cipher.None(), nil)
		caManager = builtin.NewBuiltinCaManager(secretManager)
	})

	AfterEach(func() {
		core.Now = time.Now
	})

	Context("Ensure", func() {
		It("should create a CA", func() {
			// given
			mesh := "default"
			backend := &mesh_proto.CertificateAuthorityBackend{
				Name: "builtin-1",
				Type: "builtin",
			}

			// when
			err := caManager.Ensure(context.Background(), mesh, backend)

			// then
			Expect(err).ToNot(HaveOccurred())

			// and key+cert are stored as a secrets
			secretRes := system.NewSecretResource()
			err = secretManager.Get(context.Background(), secretRes, core_store.GetByKey("default.ca-builtin-cert-builtin-1", "default"))
			Expect(err).ToNot(HaveOccurred())
			Expect(secretRes.Spec.GetData().GetValue()).ToNot(BeEmpty())

			keyRes := system.NewSecretResource()
			err = secretManager.Get(context.Background(), keyRes, core_store.GetByKey("default.ca-builtin-key-builtin-1", "default"))
			Expect(err).ToNot(HaveOccurred())
			Expect(keyRes.Spec.GetData().GetValue()).ToNot(BeEmpty())

			// when called Ensured after CA is already created
			err = caManager.Ensure(context.Background(), mesh, backend)

			// then no error happens
			Expect(err).ToNot(HaveOccurred())

			// and CA has default parameters
			block, _ := pem.Decode(secretRes.Spec.Data.Value)
			cert, err := x509.ParseCertificate(block.Bytes)
			Expect(err).ToNot(HaveOccurred())
			Expect(cert.NotAfter).To(Equal(core.Now().UTC().Add(10 * 365 * 24 * time.Hour).Truncate(time.Second)))
		})

		It("should create a configured CA", func() {
			// given
			mesh := "default"
			backend := &mesh_proto.CertificateAuthorityBackend{
				Name: "builtin-1",
				Type: "builtin",
				Conf: util_proto.MustToStruct(&config.BuiltinCertificateAuthorityConfig{
					CaCert: &config.BuiltinCertificateAuthorityConfig_CaCert{
						RSAbits:    util_proto.UInt32(uint32(2048)),
						Expiration: "1m",
					},
				}),
			}

			// when
			err := caManager.Ensure(context.Background(), mesh, backend)

			// then
			Expect(err).ToNot(HaveOccurred())

			// and CA has configured parameters
			secretRes := system.NewSecretResource()
			err = secretManager.Get(context.Background(), secretRes, core_store.GetByKey("default.ca-builtin-cert-builtin-1", "default"))
			Expect(err).ToNot(HaveOccurred())
			block, _ := pem.Decode(secretRes.Spec.Data.Value)
			cert, err := x509.ParseCertificate(block.Bytes)
			Expect(err).ToNot(HaveOccurred())
			Expect(cert.NotAfter).To(Equal(core.Now().UTC().Add(time.Minute).Truncate(time.Second)))
		})
	})

	Context("GetRootCert", func() {
		It("should retrieve created certs", func() {
			// given
			mesh := "default"
			backend := &mesh_proto.CertificateAuthorityBackend{
				Name: "builtin-1",
				Type: "builtin",
			}
			err := caManager.Ensure(context.Background(), mesh, backend)
			Expect(err).ToNot(HaveOccurred())

			// when
			certs, err := caManager.GetRootCert(context.Background(), mesh, backend)

			// then
			Expect(err).ToNot(HaveOccurred())
			Expect(certs).To(HaveLen(1))
			Expect(certs[0]).ToNot(BeEmpty())
		})

		It("should throw an error on retrieving certs on CA that was not created", func() {
			// given
			mesh := "default"
			backend := &mesh_proto.CertificateAuthorityBackend{
				Name: "builtin-non-existent",
				Type: "builtin",
			}

			// when
			_, err := caManager.GetRootCert(context.Background(), mesh, backend)

			// then
			Expect(err).To(MatchError(`failed to load CA key pair for Mesh "default" and backend "builtin-non-existent": Resource not found: type="Secret" name="default.ca-builtin-cert-builtin-non-existent" mesh="default"`))
		})
	})

	Context("GenerateDataplaneCert", func() {
		It("should generate dataplane certs", func() {
			// given
			mesh := "default"
			backend := &mesh_proto.CertificateAuthorityBackend{
				Name: "builtin-1",
				Type: "builtin",
				DpCert: &mesh_proto.CertificateAuthorityBackend_DpCert{
					Rotation: &mesh_proto.CertificateAuthorityBackend_DpCert_Rotation{
						Expiration: "1s",
					},
				},
			}
			err := caManager.Ensure(context.Background(), mesh, backend)
			Expect(err).ToNot(HaveOccurred())

			// when
			tags := map[string]map[string]bool{
				"kuma.io/service": {
					"web":     true,
					"web-api": true,
				},
				"version": {
					"v1": true,
				},
			}
			pair, err := caManager.GenerateDataplaneCert(context.Background(), mesh, backend, tags)

			// then
			Expect(err).ToNot(HaveOccurred())
			Expect(pair.KeyPEM).ToNot(BeEmpty())
			Expect(pair.CertPEM).ToNot(BeEmpty())

			// and should generate cert for dataplane with spiffe URI
			block, _ := pem.Decode(pair.CertPEM)
			cert, err := x509.ParseCertificate(block.Bytes)
			Expect(err).ToNot(HaveOccurred())
			Expect(cert.URIs).To(HaveLen(5))
			Expect(cert.URIs[0].String()).To(Equal("spiffe://default/web"))
			Expect(cert.URIs[1].String()).To(Equal("spiffe://default/web-api"))
			Expect(cert.URIs[2].String()).To(Equal("kuma://kuma.io/service/web"))
			Expect(cert.URIs[3].String()).To(Equal("kuma://kuma.io/service/web-api"))
			Expect(cert.URIs[4].String()).To(Equal("kuma://version/v1"))
			Expect(cert.NotAfter).To(Equal(now.UTC().Truncate(time.Second).Add(1 * time.Second))) // time in cert is in UTC and truncated to seconds
		})

		It("should throw an error on generate dataplane certs on non-existing CA", func() {
			// given
			mesh := "default"
			backend := &mesh_proto.CertificateAuthorityBackend{
				Name: "builtin-non-existent",
				Type: "builtin",
			}

			// when
			_, err := caManager.GenerateDataplaneCert(context.Background(), mesh, backend, mesh_proto.MultiValueTagSet{})

			// then
			Expect(err).To(MatchError(`failed to load CA key pair for Mesh "default" and backend "builtin-non-existent": Resource not found: type="Secret" name="default.ca-builtin-cert-builtin-non-existent" mesh="default"`))
		})
	})
})
