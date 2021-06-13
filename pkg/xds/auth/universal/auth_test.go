package universal_test

import (
	"context"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/pkg/tokens/builtin/zoneingress"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_store "github.com/kumahq/kuma/pkg/core/resources/store"
	"github.com/kumahq/kuma/pkg/plugins/resources/memory"
	test_model "github.com/kumahq/kuma/pkg/test/resources/model"
	builtin_issuer "github.com/kumahq/kuma/pkg/tokens/builtin/issuer"
	"github.com/kumahq/kuma/pkg/xds/auth"
	"github.com/kumahq/kuma/pkg/xds/auth/universal"
)

var _ = Describe("Authentication flow", func() {
	var privateKey = []byte("testPrivateKey")

	issuer := builtin_issuer.NewDataplaneTokenIssuer(func(string) ([]byte, error) {
		return privateKey, nil
	})
	zoneIngressIssuer := zoneingress.NewTokenIssuer(func() ([]byte, error) {
		return privateKey, nil
	})
	var authenticator auth.Authenticator
	var resStore core_store.ResourceStore

	dpRes := core_mesh.DataplaneResource{
		Meta: &test_model.ResourceMeta{
			Mesh: "dp-1",
			Name: "default",
		},
		Spec: &mesh_proto.Dataplane{
			Networking: &mesh_proto.Dataplane_Networking{
				Address: "127.0.0.1",
				Inbound: []*mesh_proto.Dataplane_Networking_Inbound{
					{
						Port:        8080,
						ServicePort: 8081,
						Tags: map[string]string{
							"kuma.io/service":  "web",
							"kuma.io/protocol": "http",
						},
					},
					{
						Port:        8090,
						ServicePort: 8091,
						Tags: map[string]string{
							"kuma.io/service":  "web-api",
							"kuma.io/protocol": "http",
						},
					},
				},
			},
		},
	}

	ingressDp := core_mesh.DataplaneResource{
		Meta: &test_model.ResourceMeta{
			Mesh: "ingress-1",
			Name: "default",
		},
		Spec: &mesh_proto.Dataplane{
			Networking: &mesh_proto.Dataplane_Networking{
				Ingress: &mesh_proto.Dataplane_Networking_Ingress{},
				Inbound: []*mesh_proto.Dataplane_Networking_Inbound{
					{
						Port:        8080,
						ServicePort: 8081,
						Tags: map[string]string{
							"kuma.io/service": "ingress",
						},
					},
				},
			},
		},
	}

	BeforeEach(func() {
		resStore = memory.NewStore()
		authenticator = universal.NewAuthenticator(issuer, zoneIngressIssuer, "zone-1")

		err := resStore.Create(context.Background(), &dpRes, core_store.CreateByKey("dp-1", "default"))
		Expect(err).ToNot(HaveOccurred())
	})

	type testCase struct {
		id    builtin_issuer.DataplaneIdentity
		dpRes *core_mesh.DataplaneResource
		err   string
	}
	DescribeTable("should correctly authenticate dataplane",
		func(given testCase) {
			// when
			credential, err := issuer.Generate(given.id)

			// then
			Expect(err).ToNot(HaveOccurred())

			// when
			err = authenticator.Authenticate(context.Background(), given.dpRes, credential)

			// then
			Expect(err).ToNot(HaveOccurred())
		},
		Entry("should auth with token bound to mesh", testCase{
			id: builtin_issuer.DataplaneIdentity{
				Mesh: "default",
			},
			dpRes: &dpRes,
		}),
		Entry("should auth with token bound to mesh and name", testCase{
			id: builtin_issuer.DataplaneIdentity{
				Name: "dp-1",
				Mesh: "default",
			},
			dpRes: &dpRes,
		}),
		Entry("should auth with token bound to mesh and tags", testCase{
			id: builtin_issuer.DataplaneIdentity{
				Mesh: "default",
				Tags: map[string]map[string]bool{
					"kuma.io/service": {
						"web":     true,
						"web-api": true,
					},
				},
			},
			dpRes: &dpRes,
		}),
		Entry("should auth with ingress token", testCase{
			id: builtin_issuer.DataplaneIdentity{
				Type: mesh_proto.IngressProxyType,
			},
			dpRes: &ingressDp,
		}),
	)

	DescribeTable("should fail auth",
		func(given testCase) {
			// when
			token, err := issuer.Generate(given.id)

			// then
			Expect(err).ToNot(HaveOccurred())

			// when
			err = authenticator.Authenticate(context.Background(), given.dpRes, token)

			// then
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring(given.err))
		},
		Entry("on token with different name", testCase{
			id: builtin_issuer.DataplaneIdentity{
				Mesh: "default",
				Name: "dp-2",
			},
			dpRes: &dpRes,
			err:   "proxy name from requestor: dp-1 is different than in token: dp-2",
		}),
		Entry("on token with different mesh", testCase{
			id: builtin_issuer.DataplaneIdentity{
				Mesh: "demo",
				Name: "dp-1",
			},
			dpRes: &dpRes,
			err:   "proxy mesh from requestor: default is different than in token: demo",
		}),
		Entry("on token with different tags", testCase{
			id: builtin_issuer.DataplaneIdentity{
				Tags: map[string]map[string]bool{
					"kuma.io/service": {
						"backend": true,
					},
				},
			},
			dpRes: &dpRes,
			err:   `which is not allowed with this token. Allowed values in token are ["backend"]`,
		}),
		Entry("on token with tag that is absent in dataplane", testCase{
			id: builtin_issuer.DataplaneIdentity{
				Tags: map[string]map[string]bool{
					"kuma.io/zone": {
						"east": true,
					},
				},
			},
			dpRes: &dpRes,
			err:   `dataplane has no tag "kuma.io/zone" required by the token`,
		}),
		Entry("on token with missing one tag value", testCase{
			id: builtin_issuer.DataplaneIdentity{
				Tags: map[string]map[string]bool{
					"kuma.io/service": {
						"web": true,
						// "web-api": true valid token should have also web-api
					},
				},
			},
			dpRes: &dpRes,
			err:   `which is not allowed with this token. Allowed values in token are ["web"]`, // web and web-api order is not stable
		}),
		Entry("regular dataplane and ingress type", testCase{
			id: builtin_issuer.DataplaneIdentity{
				Type: mesh_proto.IngressProxyType,
			},
			dpRes: &dpRes,
			err:   `dataplane is of type Dataplane but token allows only for the "ingress" type`,
		}),
		Entry("ingress dataplane and dataplane type", testCase{
			id: builtin_issuer.DataplaneIdentity{
				Type: mesh_proto.DataplaneProxyType,
			},
			dpRes: &ingressDp,
			err:   `dataplane is of type Ingress but token allows only for the "dataplane" type`,
		}),
		Entry("ingress dataplane and dataplane type (but not explicitly specified)", testCase{
			id:    builtin_issuer.DataplaneIdentity{},
			dpRes: &ingressDp,
			err:   `dataplane is of type Ingress but token allows only for the "dataplane" type`,
		}),
	)

	It("should throw an error on invalid token", func() {
		// when
		err := authenticator.Authenticate(context.Background(), &dpRes, "this-is-not-valid-jwt-token")

		// then
		Expect(err).To(MatchError("could not parse token: token contains an invalid number of segments"))
	})

	It("should throw an error when signing key is not found", func() {
		// given
		issuer := builtin_issuer.NewDataplaneTokenIssuer(func(string) ([]byte, error) {
			return nil, nil
		})

		// when
		_, err := issuer.Generate(builtin_issuer.DataplaneIdentity{
			Mesh: "demo",
		})

		// then
		Expect(err).To(MatchError(`there is no Signing Key in the Control Plane for Mesh "demo". Make sure the Mesh exist. If you run multi-zone setup, make sure Zone CP is connected to the Global before generating tokens.`))
	})
})
