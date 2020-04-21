package rest_test

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"path/filepath"

	"github.com/emicklei/go-restful"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

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

		It("should not allow to add improper CA certificate", func() {
			// when
			pair.CertPEM = []byte(`
-----BEGIN CERTIFICATE-----
MIIDKzCCAhOgAwIBAgIBADANBgkqhkiG9w0BAQsFADAwMQ0wCwYDVQQKEwRLdW1h
MQ0wCwYDVQQLEwRNZXNoMRAwDgYDVQQDEwdkZWZhdWx0MB4XDTIwMDEyOTE2MDgw
NFoXDTMwMDEyNjE2MDgxNFowQDENMAsGA1UEChMES3VtYTEdMAsGA1UECxMETWVz
aDAOBgNVBAsTB2xldmVsLTExEDAOBgNVBAMTB2RlZmF1bHQwggEiMA0GCSqGSIb3
DQEBAQUAA4IBDwAwggEKAoIBAQC+HgOFAO4T+heMQnO3VOaCYzqzkO2whggQ+yzF
0OIMSOYnOEYA93QGYB+vPf99dnoKqqp8C9MqRMMaNI3SEa23W2EXvbB41+ZBf1zW
BS6QQDSARFI6xJaXOnwDNzl+Ne2XmOJCS9pqk2hDw+WMtzxISyCd293EPW0NVY3q
H3RsyC26rIozsWf/X6fCFETEpFM5KVHSJL1511Y9w3cc0auXe+MIXShQX/v36ckm
p5pdrC3T7Y8IyTFjpPrLMVpL+HQ7LeZcs5DUfmfdxnYxIQwpBKz19/yDN4elbJem
AHVcQo1xS+KtUp9on9YTm/SuJ4bz0H9idBN25CevCY9W7CbjAgMBAAGjQDA+MA4G
A1UdDwEB/wQEAwIBBjAPBgNVHRMBAf8EBTADAQH/MBsGA1UdEQQUMBKGEHNwaWZm
ZTovL2RlZmF1bHQwDQYJKoZIhvcNAQELBQADggEBACVXnYWCCrji551pbJsOCGYJ
GEqlvcwNnnYdykas4GrfsbW2rglmaXv0uG8iH2sAH+4/MjGjnlQ6Y6Fj7mDFnidj
ugU964sEDnLuU0CtaIpHl7VZ13I0EzmfY+GsCrcIXIxbAxwWTJhz77XqbHe3baLx
Sh9wHgz/aZuy99rq9OoAvUALEaIfxrvUsVs25jLuv0Xzy57B2Dpqo0odshDA4WSS
MynQnSX7aFg1jqZQL4YjPHryEQQRj8mgjqiWp8M4/PHq5s09zDMB0DCag0QtdC/k
ydtqRoojiRS2fXY8DhFRqqRVBqLvA+7eTEKpzfjUTyEovMqxIM2n4U5MSGKQlbM=
-----END CERTIFICATE-----
			
`)
			pair.KeyPEM = []byte(`
-----BEGIN RSA PRIVATE KEY-----
MIIEowIBAAKCAQEA2F7A2cdkjR/7nm+yb6IZ6P00BgW0JvYU3MUdmUZy7QXTthID
Kis8m4z7pXcEVxjYiRYcvxY6ig2YnRf/LMF3hFKtvdNpwYq1oNrJGNOjwk5bs2WH
5a/BHXwoeMJBcOOWqSZ8TF1ghRmM6ok6G72c7n1+rPRIkW+jZYqSQ/sJpMtBe3Wn
f4ebVejlYJdl8PtbntiE0FlkuWOKn83ykMvtvZWSF763Fu/8+pWsNEwajJEGc80B
9LtI+hxFtwRL22HYJ6w274jCW27iAkE5BunJ0EiMm+vXOvH/PFvEz4RupsavqDCq
4a8pPJcerbS7kITi1bLcYpKHksRUi09LTrp3lQIDAQABAoIBADUU2d8TqbltlT9D
S9VTQWQFalPn5lCAopGe0ioePGelvFC4jooz3USUC9CGKExtzgGjqR3ACFCCEWTI
1FNYi0etOO6PBSz0KKbzxc4PbednbdvPFs3klk3zfcJSddeKHhYVWP0rE1jT8dxA
Gj9f/zYLF566t2rmpoFsw4Fl/vGscBzOq777oOPNviGPA+MSXOeI+xbMGRq+fU3R
RoMaHzjokCibQVWZb0FaLpkYGCNz7P1Zvhpkt0OZrB//e5oNjX7ksTpJsQmffCMl
XG/wX34KErjKD4CiNL5Y2CtfOKNstjS7yZoM81BIs/1dQ6OA9VGthHn6Qh8hXn4B
Y0M7wCkCgYEA97yDytk7rF6vtC3vYSKOkuYY+X06h+avDOUGB3hTYLd1vcNkA0hF
FFigof349yYG3JUdXcPPehtMzNWn1zOKfXnjHrnd/RHctuhIkKeX4AtI7+MFMT2n
vXttwBcYgRXo0+isGnSysBdHI/sF/VMLcszy5eUfy0EKxFBemufaPpMCgYEA35Zk
2DJkQ88nEtRADfHExYHJNxEmsq3V1PKQT+j68zKE84zzp1emyYH+d3ur7wXr86ZX
UA4aDVhAXcD7NUq9mVecUaRYRYmBxap1fuvu+wXXYVFsxhwWraKGSrYzwPDrt3xb
eqWAetiAqmJp8UZpyev7EHOYsWf+EUZwJLiKojcCgYBpZZCEeotCuD30YB6ZqsQR
h0dUzYxbSS9sQvufrfd7DFJRW5FvPA33rAUbJhwHuevtaJtHywi4IGk6NCPmEI14
+KRB7D2fbzwBrS1CLasVrHdpZ6JL4rk8igiVUr4gHRwjG7gswT1MYXroueFAd1ZF
jyA/4oz2QkO8ZZz6Nm3JdQKBgCIH6wt5CAfGJOVZxvIYZWHGclDeXGx/xvclgE+Z
X3DatJ+5SXCkB6/OCGQ5P58e4J3yKIH304FKeGmMsO+Yk6keS52ljQXwev8SBdYu
pO4yImkekpbIua7t+NCwUMpCIS6JUAcn35lTEKpeVk+x7vIb59fGMGx4LpSEixcb
u4YbAoGBAIgfmzZ3SCLx4pBC5/o/LdVMpzfV3vPzvu5dIsQDOat70aTHmDl+S6Cq
K96xPXFnxPS0a5TLvVCcGnA39iZDgaIWCYEEmRRsnYhlKzkNgqJEB8ZZfENwFBuO
1kXKbhap66yPSayVOAfyVS4ACia8BwT+x64AFSKjaudVNX+rGatX
-----END RSA PRIVATE KEY-----
`)
			_, err := client.AddSigningCertificate("demo", pair)

			// then
			Expect(err).To(HaveOccurred())
			apiErr := err.(*types.Error)

			Expect(*apiErr).To(Equal(types.Error{
				Title:   "Could not add signing cert",
				Details: "Resource is not valid",
				Causes: []types.Cause{
					{
						Field:   ".",
						Message: "not a valid TLS key pair: tls: private key does not match public key",
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

		XIt("should throw an error when mTLS is active and type is provided", func() { // to be removed with the WS
			// given mesh with active mTLS and type provided
			meshName := "demo"
			meshRes := &core_mesh.MeshResource{
				Spec: mesh_proto.Mesh{
					Mtls: &mesh_proto.Mesh_Mtls{
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
