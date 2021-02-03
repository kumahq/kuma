package admin_test

import (
	"context"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	kuma_cp "github.com/kumahq/kuma/pkg/config/app/kuma-cp"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	mesh_core "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	core_store "github.com/kumahq/kuma/pkg/core/resources/store"
	"github.com/kumahq/kuma/pkg/defaults/mesh"
	"github.com/kumahq/kuma/pkg/envoy/admin"
	test_model "github.com/kumahq/kuma/pkg/test/resources/model"
	"github.com/kumahq/kuma/pkg/test/runtime"
)

var _ = Describe("EnvoyAdminClient", func() {

	const (
		testMesh    = "test-mesh"
		anotherMesh = "another-mesh"
	)

	// setup the runtime
	cfg := kuma_cp.DefaultConfig()
	builder, err := runtime.BuilderFor(cfg)
	Expect(err).ToNot(HaveOccurred())
	runtime, err := builder.Build()
	Expect(err).ToNot(HaveOccurred())
	resManager := runtime.ResourceManager()
	Expect(resManager).ToNot(BeNil())

	// create mesh defaults
	err = resManager.Create(context.Background(), core_mesh.NewMeshResource(), core_store.CreateByKey(testMesh, model.NoMesh))
	Expect(err).ToNot(HaveOccurred())

	err = mesh.EnsureDefaultMeshResources(runtime.ResourceManager(), testMesh)
	Expect(err).ToNot(HaveOccurred())

	err = resManager.Create(context.Background(), core_mesh.NewMeshResource(), core_store.CreateByKey(anotherMesh, model.NoMesh))
	Expect(err).ToNot(HaveOccurred())

	err = mesh.EnsureDefaultMeshResources(runtime.ResourceManager(), anotherMesh)
	Expect(err).ToNot(HaveOccurred())

	// setup the Envoy Admin Client
	eac := admin.NewEnvoyAdminClient(resManager, runtime.Config())
	Expect(eac).ToNot(BeNil())

	Describe("GenerateAPIToken()", func() {
		It("create and fetch same token for dp/mesh", func() {
			// when
			token1, err := eac.GenerateAPIToken(&mesh_core.DataplaneResource{
				Meta: &test_model.ResourceMeta{
					Name: "dp-1",
					Mesh: testMesh,
				},
			})
			Expect(err).ToNot(HaveOccurred())

			// and
			token2, err := eac.GenerateAPIToken(&mesh_core.DataplaneResource{
				Meta: &test_model.ResourceMeta{
					Name: "dp-1",
					Mesh: testMesh,
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
					Mesh: testMesh,
				},
			})
			Expect(err).ToNot(HaveOccurred())

			// and
			token2, err := eac.GenerateAPIToken(&mesh_core.DataplaneResource{
				Meta: &test_model.ResourceMeta{
					Name: "dp-2",
					Mesh: testMesh,
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
					Mesh: testMesh,
				},
			})
			Expect(err).ToNot(HaveOccurred())

			// and
			token2, err := eac.GenerateAPIToken(&mesh_core.DataplaneResource{
				Meta: &test_model.ResourceMeta{
					Name: "dp-1",
					Mesh: anotherMesh,
				},
			})
			Expect(err).ToNot(HaveOccurred())

			// then
			Expect(token1).ToNot(Equal(token2))
		})
	})
})
