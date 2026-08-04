package model_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	mesh_proto "github.com/kumahq/kuma/v3/api/mesh/v1alpha1"
	core_mesh "github.com/kumahq/kuma/v3/pkg/core/resources/apis/mesh"
	core_model "github.com/kumahq/kuma/v3/pkg/core/resources/model"
	test_model "github.com/kumahq/kuma/v3/pkg/test/resources/model"
)

func dataplane(mesh string, name string, version string) *core_mesh.DataplaneResource {
	return &core_mesh.DataplaneResource{
		Meta: &test_model.ResourceMeta{
			Mesh:    mesh,
			Name:    name,
			Version: version,
		},
		Spec: &mesh_proto.Dataplane{},
	}
}

var _ = Describe("HashMeta", func() {
	It("should be stable for the same meta", func() {
		Expect(core_model.HashMeta(dataplane("default", "backend", "1"))).
			To(Equal(core_model.HashMeta(dataplane("default", "backend", "1"))))
	})

	It("should change when the version changes", func() {
		Expect(core_model.HashMeta(dataplane("default", "backend", "1"))).
			ToNot(Equal(core_model.HashMeta(dataplane("default", "backend", "2"))))
	})

	// the mesh and the name are written back to back, so without length prefixes
	// a resource named "bc" in mesh "a" hashes the same as "c" in mesh "ab"
	It("should not collide when the mesh and name boundary shifts", func() {
		Expect(core_model.HashMeta(dataplane("a", "bc", "1"))).
			ToNot(Equal(core_model.HashMeta(dataplane("ab", "c", "1"))))
	})
})

var _ = Describe("HashMetaIdentity", func() {
	It("should ignore the version", func() {
		Expect(core_model.HashMetaIdentity(dataplane("default", "backend", "1"))).
			To(Equal(core_model.HashMetaIdentity(dataplane("default", "backend", "2"))))
	})

	It("should not collide when the mesh and name boundary shifts", func() {
		Expect(core_model.HashMetaIdentity(dataplane("a", "bc", "1"))).
			ToNot(Equal(core_model.HashMetaIdentity(dataplane("ab", "c", "1"))))
	})
})
