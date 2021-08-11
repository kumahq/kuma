package admin_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	test_model "github.com/kumahq/kuma/pkg/test/resources/model"
)

var _ = Describe("EnvoyAdminClient", func() {

	Describe("GenerateAPIToken()", func() {
		It("create and fetch same token for dp/mesh", func() {
			// when
			token1, err := eac.GenerateAPIToken(&core_mesh.DataplaneResource{
				Meta: &test_model.ResourceMeta{
					Name: "dp-1",
					Mesh: testMesh,
				},
			})
			Expect(err).ToNot(HaveOccurred())

			// and
			token2, err := eac.GenerateAPIToken(&core_mesh.DataplaneResource{
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
			token1, err := eac.GenerateAPIToken(&core_mesh.DataplaneResource{
				Meta: &test_model.ResourceMeta{
					Name: "dp-1",
					Mesh: testMesh,
				},
			})
			Expect(err).ToNot(HaveOccurred())

			// and
			token2, err := eac.GenerateAPIToken(&core_mesh.DataplaneResource{
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
			token1, err := eac.GenerateAPIToken(&core_mesh.DataplaneResource{
				Meta: &test_model.ResourceMeta{
					Name: "dp-1",
					Mesh: testMesh,
				},
			})
			Expect(err).ToNot(HaveOccurred())

			// and
			token2, err := eac.GenerateAPIToken(&core_mesh.DataplaneResource{
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
