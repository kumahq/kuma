package rest_test

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"path/filepath"

	mesh_proto "github.com/Kong/kuma/api/mesh/v1alpha1"
	"github.com/Kong/kuma/app/kumactl/pkg/ca"
	"github.com/Kong/kuma/pkg/core/ca/provided"
	"github.com/Kong/kuma/pkg/core/ca/provided/rest"
	core_mesh "github.com/Kong/kuma/pkg/core/resources/apis/mesh"
	resources_manager "github.com/Kong/kuma/pkg/core/resources/manager"
	core_store "github.com/Kong/kuma/pkg/core/resources/store"
	"github.com/Kong/kuma/pkg/core/rest/errors/types"
	"github.com/Kong/kuma/pkg/core/secrets/cipher"
	"github.com/Kong/kuma/pkg/core/secrets/manager"
	"github.com/Kong/kuma/pkg/core/secrets/store"
	"github.com/Kong/kuma/pkg/plugins/resources/memory"
	"github.com/Kong/kuma/pkg/tls"
	"github.com/emicklei/go-restful"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Provided CA WS", func() {

	var client ca.ProvidedCaClient
	var srv *httptest.Server
	var resManager resources_manager.ResourceManager

	BeforeEach(func() {
		memStore := memory.NewStore()
		resManager = resources_manager.NewResourceManager(memStore)
		ws := rest.NewWebservice(provided.NewProvidedCaManager(manager.NewSecretManager(store.NewSecretStore(memStore), cipher.None())), resManager)
		container := restful.NewContainer()
		container.Add(ws)
		srv = httptest.NewServer(container)

		// wait for the server
		Eventually(func() error {
			_, err := http.DefaultClient.Get(fmt.Sprintf("%s/meshes/default/ca/provided", srv.URL))
			return err
		}).ShouldNot(HaveOccurred())

		c, err := ca.NewProvidedCaClient(srv.URL, nil)
		Expect(err).ToNot(HaveOccurred())
		client = c
	})

	AfterEach(func() {
		srv.Close()
	})

	var pair tls.KeyPair

	BeforeEach(func() {
		cert, err := ioutil.ReadFile(filepath.Join("testdata", "cert.pem"))
		Expect(err).ToNot(HaveOccurred())
		key, err := ioutil.ReadFile(filepath.Join("testdata", "cert.key"))
		Expect(err).ToNot(HaveOccurred())
		pair = tls.KeyPair{
			CertPEM: cert,
			KeyPEM:  key,
		}
	})

	Describe("Add signing certificate", func() {
		It("should add certificate and retrieve it", func() {
			// when
			signingCert, err := client.AddSigningCertificate("demo", pair)

			// then
			Expect(err).ToNot(HaveOccurred())

			// when
			certs, err := client.SigningCertificates("demo")

			// then
			Expect(err).ToNot(HaveOccurred())

			// and
			Expect(certs).To(HaveLen(1))
			Expect(certs[0]).To(Equal(signingCert))
		})

		It("should not allow to add certificate without key", func() {
			// when
			pair.KeyPEM = []byte{}
			_, err := client.AddSigningCertificate("demo", pair)

			// then
			Expect(err).To(HaveOccurred())
			apiErr := err.(*types.Error)

			Expect(*apiErr).To(Equal(types.Error{
				Title:   "Could not add signing cert",
				Details: "Resource is not valid",
				Causes: []types.Cause{
					{
						Field:   "key",
						Message: "must not be empty",
					},
				},
			}))
		})

		It("should not allow to add certificate without cert", func() {
			// when
			pair.CertPEM = []byte{}
			_, err := client.AddSigningCertificate("demo", pair)

			// then
			Expect(err).To(HaveOccurred())
			apiErr := err.(*types.Error)

			Expect(*apiErr).To(Equal(types.Error{
				Title:   "Could not add signing cert",
				Details: "Resource is not valid",
				Causes: []types.Cause{
					{
						Field:   "cert",
						Message: "must not be empty",
					},
				},
			}))
		})
	})

	Describe("Get certificates", func() {
		It("should return error when there are no certs", func() {
			// when
			_, err := client.SigningCertificates("non-existing-mesh")

			// then
			Expect(err).To(HaveOccurred())

			// and
			Expect(err).To(Equal(&types.Error{
				Title:   "Could not retrieve signing certs",
				Details: "Not found",
			}))
		})
	})

	Describe("Delete signing certificate", func() {
		It("should delete existing certificate", func() {
			// given
			signingCert, err := client.AddSigningCertificate("demo", pair)
			Expect(err).ToNot(HaveOccurred())

			// when
			err = client.DeleteSigningCertificate("demo", signingCert.Id)

			// then
			Expect(err).ToNot(HaveOccurred())

			// when
			certs, err := client.SigningCertificates("demo")

			// then
			Expect(err).ToNot(HaveOccurred())

			// and
			Expect(certs).To(BeEmpty())
		})

		It("should throw an error on deleting non existing certificate", func() {
			// given
			_, err := client.AddSigningCertificate("demo", pair)
			Expect(err).ToNot(HaveOccurred())

			// when
			err = client.DeleteSigningCertificate("demo", "non-existing-cert-id")

			// then
			Expect(err).To(HaveOccurred())
			Expect(err).To(Equal(&types.Error{
				Title:   "Could not delete signing cert",
				Details: "Not found",
			}))
		})

		It("should throw an error when mTLS is active and type is provided", func() {
			// given mesh with active mTLS and type provided
			meshName := "demo"
			meshRes := &core_mesh.MeshResource{
				Spec: mesh_proto.Mesh{
					Mtls: &mesh_proto.Mesh_Mtls{
						Ca: &mesh_proto.CertificateAuthority{
							Type: &mesh_proto.CertificateAuthority_Provided_{
								Provided: &mesh_proto.CertificateAuthority_Provided{},
							},
						},
						Enabled: true,
					},
				},
			}
			err := resManager.Create(context.Background(), meshRes, core_store.CreateByKey(meshName, meshName))
			Expect(err).ToNot(HaveOccurred())

			// given
			signingCert, err := client.AddSigningCertificate(meshName, pair)
			Expect(err).ToNot(HaveOccurred())

			// when
			err = client.DeleteSigningCertificate(meshName, signingCert.Id)

			// then
			Expect(err).To(HaveOccurred())
			Expect(err).To(Equal(&types.Error{
				Title:   "Could not delete signing cert",
				Details: "Cannot delete signing certificate when the mTLS in the mesh is active and type of CA is Provided",
			}))
		})
	})
})
