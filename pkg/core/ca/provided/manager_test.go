package provided_test

import (
	"context"
	"github.com/Kong/kuma/pkg/core/ca/provided"
	core_store "github.com/Kong/kuma/pkg/core/resources/store"
	"github.com/Kong/kuma/pkg/core/secrets/cipher"
	"github.com/Kong/kuma/pkg/core/secrets/manager"
	"github.com/Kong/kuma/pkg/core/secrets/store"
	"github.com/Kong/kuma/pkg/plugins/resources/memory"
	"github.com/Kong/kuma/pkg/tls"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("CA Provided Manager", func() {

	var caManager provided.ProvidedCaManager
	const meshName = "demo"

	BeforeEach(func() {
		caManager = provided.NewProvidedCaManager(manager.NewSecretManager(store.NewSecretStore(memory.NewStore()), cipher.None()))
	})

	Describe("AddSigningCert", func() {
		It("should create CA when adding new CA Root to it", func() {
			// when
			_, err := caManager.GetSigningCerts(context.Background(), meshName)

			// then
			Expect(err).To(HaveOccurred())
			Expect(core_store.IsResourceNotFound(err)).To(BeTrue())

			// when
			pair := tls.KeyPair{
				CertPEM: []byte("CERT"),
				KeyPEM:  []byte("KEY"),
			}
			signingCert, err := caManager.AddSigningCert(context.Background(), meshName, pair)

			// then
			Expect(err).ToNot(HaveOccurred())

			// and
			rootCerts, err := caManager.GetSigningCerts(context.Background(), meshName)

			// then
			Expect(err).ToNot(HaveOccurred())
			Expect(rootCerts).To(HaveLen(1))
			Expect(rootCerts[0]).To(Equal(signingCert))
		})

		It("should not allow to add another CA Root to existing CA", func() {
			// setup CA with CA Root
			caRoot := tls.KeyPair{
				CertPEM: []byte("CERT"),
				KeyPEM:  []byte("KEY"),
			}
			_, err := caManager.AddSigningCert(context.Background(), meshName, caRoot)
			Expect(err).ToNot(HaveOccurred())

			// given
			newRoot := tls.KeyPair{
				CertPEM: []byte("CERT2"),
				KeyPEM:  []byte("KEY2"),
			}

			// when
			_, err = caManager.AddSigningCert(context.Background(), meshName, newRoot)

			// then
			Expect(err).To(HaveOccurred())
			Expect(err).To(MatchError("cannot add more than 1 CA root to provided CA"))
		})
	})

	Describe("DeleteSigningCert", func() {
		BeforeEach(func() {
			// setup CA with CA Root
			caRoot := tls.KeyPair{
				CertPEM: []byte("CERT"),
				KeyPEM:  []byte("KEY"),
			}
			_, err := caManager.AddSigningCert(context.Background(), meshName, caRoot)
			Expect(err).ToNot(HaveOccurred())
		})

		It("should delete CA root", func() {
			// when
			caRootCerts, err := caManager.GetSigningCerts(context.Background(), meshName)

			// then
			Expect(err).ToNot(HaveOccurred())
			Expect(caRootCerts).To(HaveLen(1))

			// when
			err = caManager.DeleteSigningCert(context.Background(), meshName, caRootCerts[0].Id)

			// then
			Expect(err).ToNot(HaveOccurred())

			// when
			caRootCerts, err = caManager.GetSigningCerts(context.Background(), meshName)

			// then
			Expect(err).ToNot(HaveOccurred())
			Expect(caRootCerts).To(HaveLen(0))
		})

		It("should throw an error for invalid mesh", func() {
			// when
			err := caManager.DeleteSigningCert(context.Background(), "unknown-mesh", "id-1")

			// then
			Expect(err).To(HaveOccurred())
			Expect(err).To(MatchError(`failed to load CA key pair for Mesh "unknown-mesh": Resource not found: type="Secret" name="providedca.unknown-mesh" mesh="unknown-mesh"`))
		})

		It("should throw an error for unknown CA root", func() {
			// when
			err := caManager.DeleteSigningCert(context.Background(), meshName, "unknown-id")

			// then
			Expect(err).To(HaveOccurred())
			Expect(err).To(MatchError(`could not find CA Root of id "unknown-id" for mesh "demo"`))
		})
	})

	Describe("DeleteCa", func() {
		It("should delete CA", func() {
			// setup CA with CA Root
			caRoot := tls.KeyPair{
				CertPEM: []byte("CERT"),
				KeyPEM:  []byte("KEY"),
			}
			_, err := caManager.AddSigningCert(context.Background(), meshName, caRoot)
			Expect(err).ToNot(HaveOccurred())

			// when
			err = caManager.DeleteCa(context.Background(), meshName)

			// then
			Expect(err).ToNot(HaveOccurred())

			// when
			_, err = caManager.GetSigningCerts(context.Background(), meshName)

			// then
			Expect(err).To(HaveOccurred())
			Expect(core_store.IsResourceNotFound(err)).To(BeTrue())
		})

		It("should throw an error when deleting non existing CA", func() {
			// when
			err := caManager.DeleteCa(context.Background(), "unknown-mesh")

			// then
			Expect(err).To(HaveOccurred())
			Expect(core_store.IsResourceNotFound(err)).To(BeTrue())
		})
	})

	Describe("GenerateWorkloadCert", func() {
		BeforeEach(func() {
			// setup CA with CA Root
			pair, err := tls.NewSelfSignedCert("kuma", tls.ServerCertType)
			Expect(err).ToNot(HaveOccurred())
			_, err = caManager.AddSigningCert(context.Background(), meshName, pair)
			Expect(err).ToNot(HaveOccurred())
		})

		It("should generate workload cert", func() {
			// when
			pair, err := caManager.GenerateWorkloadCert(context.Background(), meshName, "backend")

			// then
			Expect(err).ToNot(HaveOccurred())

			// and
			Expect(pair.CertPEM).ToNot(HaveLen(0))
			Expect(pair.KeyPEM).ToNot(HaveLen(0))
		})

		It("should throw an error for mesh without CA", func() {
			// when
			_, err := caManager.GenerateWorkloadCert(context.Background(), "mesh-without-ca", "backend")

			// then
			Expect(err).To(HaveOccurred())
			Expect(err).To(MatchError(`failed to load CA key pair for Mesh "mesh-without-ca": Resource not found: type="Secret" name="providedca.mesh-without-ca" mesh="mesh-without-ca"`))
		})
	})
})
