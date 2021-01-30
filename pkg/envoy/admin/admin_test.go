package admin_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	kuma_cp "github.com/kumahq/kuma/pkg/config/app/kuma-cp"
	mesh_core "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/envoy/admin"
	test_model "github.com/kumahq/kuma/pkg/test/resources/model"
	"github.com/kumahq/kuma/pkg/test/runtime"
)

var _ = Describe("EnvoyAdminClient", func() {

	var eac admin.EnvoyAdminClient

	BeforeEach(func() {
		cfg := kuma_cp.DefaultConfig()
		builder, err := runtime.BuilderFor(cfg)
		Expect(err).ToNot(HaveOccurred())
		runtime, err := builder.Build()
		Expect(err).ToNot(HaveOccurred())
		eac = admin.NewEnvoyAdminClient(runtime.ResourceManager(), runtime.Config())
	})

	Describe("GenerateAPIToken()", func() {
		It("create and fetch same token for dp/mesh", func() {
			// when
			token1, err := eac.GenerateAPIToken(&mesh_core.DataplaneResource{
				Meta: &test_model.ResourceMeta{
					Name: "dp-1",
					Mesh: "mesh-1",
				},
			})
			Expect(err).ToNot(HaveOccurred())

			// and
			token2, err := eac.GenerateAPIToken(&mesh_core.DataplaneResource{
				Meta: &test_model.ResourceMeta{
					Name: "dp-1",
					Mesh: "mesh-1",
				},
			})
			Expect(err).ToNot(HaveOccurred())

			// then
			Expect(token1).To(Equal(token2))
		})

		It("two dps in same mesh", func() {
			// when
			token1, err := eac.GenerateAPIToken(&mesh_core.DataplaneResource{
				Meta: &test_model.ResourceMeta{
					Name: "dp-1",
					Mesh: "mesh-1",
				},
			})
			Expect(err).ToNot(HaveOccurred())

			// and
			token2, err := eac.GenerateAPIToken(&mesh_core.DataplaneResource{
				Meta: &test_model.ResourceMeta{
					Name: "dp-2",
					Mesh: "mesh-1",
				},
			})
			Expect(err).ToNot(HaveOccurred())

			// then
			Expect(token1).ToNot(Equal(token2))
		})

		It("two dps in two meshes", func() {
			// when
			token1, err := eac.GenerateAPIToken(&mesh_core.DataplaneResource{
				Meta: &test_model.ResourceMeta{
					Name: "dp-1",
					Mesh: "mesh-1",
				},
			})
			Expect(err).ToNot(HaveOccurred())

			// and
			token2, err := eac.GenerateAPIToken(&mesh_core.DataplaneResource{
				Meta: &test_model.ResourceMeta{
					Name: "dp-1",
					Mesh: "mesh-2",
				},
			})
			Expect(err).ToNot(HaveOccurred())

			// then
			Expect(token1).ToNot(Equal(token2))
		})
	})
})
