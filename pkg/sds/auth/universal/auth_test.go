package universal_test

import (
	"context"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/manager"
	"github.com/kumahq/kuma/pkg/core/resources/store"
	"github.com/kumahq/kuma/pkg/core/xds"
	"github.com/kumahq/kuma/pkg/plugins/resources/memory"
	"github.com/kumahq/kuma/pkg/sds/auth"
	"github.com/kumahq/kuma/pkg/sds/auth/universal"
	"github.com/kumahq/kuma/pkg/sds/server"
	builtin_issuer "github.com/kumahq/kuma/pkg/tokens/builtin/issuer"
)

var _ = Describe("Authentication flow", func() {
	var privateKey = []byte("testPrivateKey")

	issuer := builtin_issuer.NewDataplaneTokenIssuer(func() ([]byte, error) {
		return privateKey, nil
	})
	var authenticator auth.Authenticator
	var resStore store.ResourceStore

	BeforeEach(func() {
		resStore = memory.NewStore()
		authenticator = universal.NewAuthenticator(
			issuer,
			server.DefaultDataplaneResolver(manager.NewResourceManager(resStore)),
		)

		dpRes := core_mesh.DataplaneResource{
			Spec: v1alpha1.Dataplane{
				Networking: &v1alpha1.Dataplane_Networking{
					Address: "127.0.0.1",
					Inbound: []*v1alpha1.Dataplane_Networking_Inbound{
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
		err := resStore.Create(context.Background(), &dpRes, store.CreateByKey("dp-1", "default"))
		Expect(err).ToNot(HaveOccurred())

		dpRes = core_mesh.DataplaneResource{
			Spec: v1alpha1.Dataplane{
				Networking: &v1alpha1.Dataplane_Networking{
					Ingress: &v1alpha1.Dataplane_Networking_Ingress{},
					Inbound: []*v1alpha1.Dataplane_Networking_Inbound{
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
		err = resStore.Create(context.Background(), &dpRes, store.CreateByKey("ingress-1", "default"))
		Expect(err).ToNot(HaveOccurred())
	})

	type testCase struct {
		id      builtin_issuer.DataplaneIdentity
		proxyId xds.ProxyId
		service string
		err     string
	}
	DescribeTable("should correctly authenticate dataplane",
		func(given testCase) {
			// when
			credential, err := issuer.Generate(given.id)

			// then
			Expect(err).ToNot(HaveOccurred())

			// when
			authIdentity, err := authenticator.Authenticate(context.Background(), given.proxyId, auth.Credential(credential))

			// then
			Expect(err).ToNot(HaveOccurred())
			Expect(authIdentity.Services[0]).To(Equal(given.service))
			Expect(authIdentity.Mesh).To(Equal("default"))
		},
		Entry("should auth with token bound to nothing", testCase{
			id: builtin_issuer.DataplaneIdentity{
				Name: "",
				Mesh: "",
				Tags: nil,
			},
			proxyId: xds.ProxyId{
				Mesh: "default",
				Name: "dp-1",
			},
			service: "web",
		}),
		Entry("should auth with token bound to mesh", testCase{
			id: builtin_issuer.DataplaneIdentity{
				Mesh: "default",
			},
			proxyId: xds.ProxyId{
				Mesh: "default",
				Name: "dp-1",
			},
			service: "web",
		}),
		Entry("should auth with token bound to mesh and name", testCase{
			id: builtin_issuer.DataplaneIdentity{
				Name: "dp-1",
				Mesh: "default",
			},
			proxyId: xds.ProxyId{
				Mesh: "default",
				Name: "dp-1",
			},
			service: "web",
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
			proxyId: xds.ProxyId{
				Mesh: "default",
				Name: "dp-1",
			},
			service: "web",
		}),
		Entry("should auth with ingress token", testCase{
			id: builtin_issuer.DataplaneIdentity{
				Type: builtin_issuer.DpTypeIngress,
			},
			proxyId: xds.ProxyId{
				Mesh: "default",
				Name: "ingress-1",
			},
			service: "ingress",
		}),
	)

	DescribeTable("should fail auth",
		func(given testCase) {
			// when
			token, err := issuer.Generate(given.id)

			// then
			Expect(err).ToNot(HaveOccurred())

			// when
			_, err = authenticator.Authenticate(context.Background(), given.proxyId, auth.Credential(token))

			// then
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring(given.err))
		},
		Entry("on token with different name", testCase{
			id: builtin_issuer.DataplaneIdentity{
				Mesh: "default",
				Name: "dp-2",
			},
			proxyId: xds.ProxyId{
				Mesh: "default",
				Name: "dp-1",
			},
			err: "proxy name from requestor: dp-1 is different than in token: dp-2",
		}),
		Entry("on token with different mesh", testCase{
			id: builtin_issuer.DataplaneIdentity{
				Mesh: "demo",
				Name: "dp-1",
			},
			proxyId: xds.ProxyId{
				Mesh: "default",
				Name: "dp-1",
			},
			err: "proxy mesh from requestor: default is different than in token: demo",
		}),
		Entry("on token with different tags", testCase{
			id: builtin_issuer.DataplaneIdentity{
				Tags: map[string]map[string]bool{
					"kuma.io/service": {
						"backend": true,
					},
				},
			},
			proxyId: xds.ProxyId{
				Mesh: "default",
				Name: "dp-1",
			},
			err: `which is not allowed with this token. Allowed values in token are ["backend"]`,
		}),
		Entry("on token with tag that is absent in dataplane", testCase{
			id: builtin_issuer.DataplaneIdentity{
				Tags: map[string]map[string]bool{
					"kuma.io/zone": {
						"east": true,
					},
				},
			},
			proxyId: xds.ProxyId{
				Mesh: "default",
				Name: "dp-1",
			},
			err: `dataplane has no tag "kuma.io/zone" required by the token`,
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
			proxyId: xds.ProxyId{
				Mesh: "default",
				Name: "dp-1",
			},
			err: `which is not allowed with this token. Allowed values in token are ["web"]`, // web and web-api order is not stable
		}),
		Entry("regular dataplane and ingress type", testCase{
			id: builtin_issuer.DataplaneIdentity{
				Type: builtin_issuer.DpTypeIngress,
			},
			proxyId: xds.ProxyId{
				Mesh: "default",
				Name: "dp-1",
			},
			err: `dataplane is of type Dataplane but token allows for the "ingress" type`,
		}),
		Entry("ingress dataplane and dataplane type", testCase{
			id: builtin_issuer.DataplaneIdentity{
				Type: builtin_issuer.DpTypeDataplane,
			},
			proxyId: xds.ProxyId{
				Mesh: "default",
				Name: "ingress-1",
			},
			err: `dataplane is of type Ingress but token allows for the "dataplane" type`,
		}),
		Entry("ingress dataplane and dataplane type (but not explicitly specified)", testCase{
			id: builtin_issuer.DataplaneIdentity{
			},
			proxyId: xds.ProxyId{
				Mesh: "default",
				Name: "ingress-1",
			},
			err: `dataplane is of type Ingress but token allows for the "dataplane" type`,
		}),
	)

	It("should throw an error on invalid token", func() {
		// when
		id := xds.ProxyId{
			Mesh: "default",
			Name: "dp-1",
		}
		_, err := authenticator.Authenticate(context.Background(), id, "this-is-not-valid-jwt-token")

		// then
		Expect(err).To(MatchError("could not parse token: token contains an invalid number of segments"))
	})

	It("should throw an error when dataplane is not present in CP", func() {
		// given
		id := builtin_issuer.DataplaneIdentity{
			Mesh: "default",
			Name: "non-existent-dp",
		}

		// when
		token, err := issuer.Generate(id)

		// then
		Expect(err).ToNot(HaveOccurred())

		// when
		_, err = authenticator.Authenticate(context.Background(), xds.ProxyId{
			Mesh: "default",
			Name: "non-existent-dp",
		}, auth.Credential(token))

		// then
		Expect(err).To(MatchError(`unable to find Dataplane for proxy "default.non-existent-dp": Resource not found: type="Dataplane" name="non-existent-dp" mesh="default"`))
	})

	It("should throw an error when signing key is not found", func() {
		// given
		issuer := builtin_issuer.NewDataplaneTokenIssuer(func() ([]byte, error) {
			return nil, nil
		})

		// when
		_, err := issuer.Generate(builtin_issuer.DataplaneIdentity{})

		// then
		Expect(err).To(MatchError("there is no Signing Key in the Control Plane. If you run multi-zone setup, make sure Remote is connected to the Global before generating tokens."))
	})
})
