package universal_test

import (
	"context"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

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

	issuer := builtin_issuer.NewDataplaneTokenIssuer(func() ([]byte, error) {
		return privateKey, nil
	})
	var authenticator auth.Authenticator
	var resStore core_store.ResourceStore

	dpRes := core_mesh.DataplaneResource{
		Meta: &test_model.ResourceMeta{
			Mesh: "dp-1",
			Name: "default",
		},
		Spec: mesh_proto.Dataplane{
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

	BeforeEach(func() {
		resStore = memory.NewStore()
		authenticator = universal.NewAuthenticator(issuer)

		err := resStore.Create(context.Background(), &dpRes, core_store.CreateByKey("dp-1", "default"))
		Expect(err).ToNot(HaveOccurred())
	})

	DescribeTable("should correctly authenticate dataplane",
		func(id builtin_issuer.DataplaneIdentity) {
			// when
			credential, err := issuer.Generate(id)

			// then
			Expect(err).ToNot(HaveOccurred())

			// when
			err = authenticator.Authenticate(context.Background(), &dpRes, credential)

			// then
			Expect(err).ToNot(HaveOccurred())
		},
		Entry("should auth with token bound to nothing", builtin_issuer.DataplaneIdentity{
			Name: "",
			Mesh: "",
			Tags: nil,
		}),
		Entry("should auth with token bound to mesh", builtin_issuer.DataplaneIdentity{
			Mesh: "default",
		}),
		Entry("should auth with token bound to mesh and name", builtin_issuer.DataplaneIdentity{
			Name: "dp-1",
			Mesh: "default",
		}),
		Entry("should auth with token bound to mesh and tags", builtin_issuer.DataplaneIdentity{
			Mesh: "default",
			Tags: map[string]map[string]bool{
				"kuma.io/service": {
					"web":     true,
					"web-api": true,
				},
			},
		}),
	)

	type testCase struct {
		id  builtin_issuer.DataplaneIdentity
		err string
	}
	DescribeTable("should fail auth",
		func(given testCase) {
			// when
			token, err := issuer.Generate(given.id)

			// then
			Expect(err).ToNot(HaveOccurred())

			// when
			err = authenticator.Authenticate(context.Background(), &dpRes, token)

			// then
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring(given.err))
		},
		Entry("on token with different name", testCase{
			id: builtin_issuer.DataplaneIdentity{
				Mesh: "default",
				Name: "dp-2",
			},
			err: "proxy name from requestor: dp-1 is different than in token: dp-2",
		}),
		Entry("on token with different mesh", testCase{
			id: builtin_issuer.DataplaneIdentity{
				Mesh: "demo",
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
			err: `which is not allowed with this token. Allowed values in token are ["web"]`, // web and web-api order is not stable
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
		issuer := builtin_issuer.NewDataplaneTokenIssuer(func() ([]byte, error) {
			return nil, nil
		})

		// when
		_, err := issuer.Generate(builtin_issuer.DataplaneIdentity{})

		// then
		Expect(err).To(MatchError("there is no Signing Key in the Control Plane. If you run multi-zone setup, make sure Remote is connected to the Global before generating tokens."))
	})
})
