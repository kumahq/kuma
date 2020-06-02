package builtin_test

import (
	"context"
	"crypto/x509"
	"encoding/pem"
	"time"

	"github.com/golang/protobuf/ptypes/wrappers"

	mesh_proto "github.com/Kong/kuma/api/mesh/v1alpha1"
	"github.com/Kong/kuma/pkg/core"
	core_ca "github.com/Kong/kuma/pkg/core/ca"
	"github.com/Kong/kuma/pkg/core/resources/apis/system"
	core_store "github.com/Kong/kuma/pkg/core/resources/store"
	"github.com/Kong/kuma/pkg/core/secrets/cipher"
	secret_manager "github.com/Kong/kuma/pkg/core/secrets/manager"
	"github.com/Kong/kuma/pkg/core/secrets/store"
	"github.com/Kong/kuma/pkg/plugins/ca/builtin"
	"github.com/Kong/kuma/pkg/plugins/ca/builtin/config"
	"github.com/Kong/kuma/pkg/plugins/resources/memory"
	"github.com/Kong/kuma/pkg/util/proto"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Builtin CA Manager", func() {

	var secretManager secret_manager.SecretManager
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
			//given
			mesh := "default"
			backend := mesh_proto.CertificateAuthorityBackend{
				Name: "builtin-1",
				Type: "builtin",
			}

			// when
			err := caManager.Ensure(context.Background(), mesh, backend)

			// then
			Expect(err).ToNot(HaveOccurred())

			// and key+cert are stored as a secrets
			secretRes := system.SecretResource{}
			err = secretManager.Get(context.Background(), &secretRes, core_store.GetByKey("default.ca-builtin-cert-builtin-1", "default"))
			Expect(err).ToNot(HaveOccurred())
			Expect(secretRes.Spec.GetData().GetValue()).ToNot(BeEmpty())

			keyRes := system.SecretResource{}
			err = secretManager.Get(context.Background(), &keyRes, core_store.GetByKey("default.ca-builtin-key-builtin-1", "default"))
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
			//given
			mesh := "default"
			backend := mesh_proto.CertificateAuthorityBackend{
				Name: "builtin-1",
				Type: "builtin",
				Conf: proto.MustToStruct(&config.BuiltinCertificateAuthorityConfig{
					CaCert: &config.BuiltinCertificateAuthorityConfig_CaCert{
						RSAbits: &wrappers.UInt32Value{
							Value: uint32(2048),
						},
						Expiration: "1m",
					},
				}),
			}

			// when
			err := caManager.Ensure(context.Background(), mesh, backend)

			// then
			Expect(err).ToNot(HaveOccurred())

			// and CA has configured parameters
			secretRes := system.SecretResource{}
			err = secretManager.Get(context.Background(), &secretRes, core_store.GetByKey("default.ca-builtin-cert-builtin-1", "default"))
			Expect(err).ToNot(HaveOccurred())
			block, _ := pem.Decode(secretRes.Spec.Data.Value)
			cert, err := x509.ParseCertificate(block.Bytes)
			Expect(err).ToNot(HaveOccurred())
			Expect(cert.NotAfter).To(Equal(core.Now().UTC().Add(time.Minute).Truncate(time.Second)))
		})
	})

	Context("GetRootCert", func() {
		It("should retrieve created certs", func() {
			//given
			mesh := "default"
			backend := mesh_proto.CertificateAuthorityBackend{
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
			backend := mesh_proto.CertificateAuthorityBackend{
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
			//given
			mesh := "default"
			backend := mesh_proto.CertificateAuthorityBackend{
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
			pair, err := caManager.GenerateDataplaneCert(context.Background(), mesh, backend, []string{"web", "web-api"})

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

		It("should throw an error on generate dataplane certs on non-existing CA", func() {
			// given
			mesh := "default"
			backend := mesh_proto.CertificateAuthorityBackend{
				Name: "builtin-non-existent",
				Type: "builtin",
			}

			// when
			_, err := caManager.GenerateDataplaneCert(context.Background(), mesh, backend, []string{"web"})

			// then
			Expect(err).To(MatchError(`failed to load CA key pair for Mesh "default" and backend "builtin-non-existent": Resource not found: type="Secret" name="default.ca-builtin-cert-builtin-non-existent" mesh="default"`))
		})
	})
})
