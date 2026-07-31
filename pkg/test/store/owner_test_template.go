package store

import (
	"context"
	"fmt"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	core_meta "github.com/kumahq/kuma/v3/pkg/core/metadata"
	core_mesh "github.com/kumahq/kuma/v3/pkg/core/resources/apis/mesh"
	meshexternalservice_api "github.com/kumahq/kuma/v3/pkg/core/resources/apis/meshexternalservice/api/v1alpha1"
	secret_model "github.com/kumahq/kuma/v3/pkg/core/resources/apis/system"
	"github.com/kumahq/kuma/v3/pkg/core/resources/model"
	"github.com/kumahq/kuma/v3/pkg/core/resources/store"
)

func ExecuteOwnerTests(
	createStore func() store.ResourceStore,
	storeName string,
) {
	const mesh = "default-mesh"
	var s store.ClosableResourceStore

	BeforeEach(func() {
		s = store.NewStrictResourceStore(createStore())
	})

	AfterEach(func() {
		err := s.Close()
		Expect(err).ToNot(HaveOccurred())
	})

	newMeshExternalService := func() *meshexternalservice_api.MeshExternalServiceResource {
		return &meshexternalservice_api.MeshExternalServiceResource{
			Spec: &meshexternalservice_api.MeshExternalService{
				Match: meshexternalservice_api.Match{
					Type:     meshexternalservice_api.HostnameGeneratorType,
					Port:     80,
					Protocol: core_meta.ProtocolHTTP,
				},
				Endpoints: &[]meshexternalservice_api.Endpoint{{Address: "demo.example.com", Port: 80}},
			},
			Status: &meshexternalservice_api.MeshExternalServiceStatus{},
		}
	}

	Context("Store: "+storeName, func() {
		It("should delete secret when its owner is deleted", func() {
			// setup
			meshRes := core_mesh.NewMeshResource()
			err := s.Create(context.Background(), meshRes, store.CreateByKey(mesh, model.NoMesh))
			Expect(err).ToNot(HaveOccurred())

			name := "secret-1"
			secretRes := secret_model.NewSecretResource()
			err = s.Create(context.Background(), secretRes,
				store.CreateByKey(name, mesh),
				store.CreatedAt(time.Now()),
				store.CreateWithOwner(meshRes))
			Expect(err).ToNot(HaveOccurred())

			// when
			err = s.Delete(context.Background(), meshRes, store.DeleteByKey(mesh, model.NoMesh))
			Expect(err).ToNot(HaveOccurred())

			// then
			actual := secret_model.NewSecretResource()
			err = s.Get(context.Background(), actual, store.GetByKey(name, mesh))
			Expect(store.IsNotFound(err)).To(BeTrue())
		})

		It("should delete resource when its owner is deleted", func() {
			// setup
			meshRes := core_mesh.NewMeshResource()
			err := s.Create(context.Background(), meshRes, store.CreateByKey(mesh, model.NoMesh))
			Expect(err).ToNot(HaveOccurred())

			name := "resource-1"
			trRes := newMeshExternalService()
			err = s.Create(context.Background(), trRes,
				store.CreateByKey(name, mesh),
				store.CreatedAt(time.Now()),
				store.CreateWithOwner(meshRes))
			Expect(err).ToNot(HaveOccurred())

			// when
			err = s.Delete(context.Background(), meshRes, store.DeleteByKey(mesh, model.NoMesh))
			Expect(err).ToNot(HaveOccurred())

			// then
			actual := meshexternalservice_api.NewMeshExternalServiceResource()
			err = s.Get(context.Background(), actual, store.GetByKey(name, mesh))
			Expect(store.IsNotFound(err)).To(BeTrue())
		})

		It("should delete resource when its owner is deleted after owner update", func() {
			// setup
			meshRes := core_mesh.NewMeshResource()
			err := s.Create(context.Background(), meshRes, store.CreateByKey(mesh, model.NoMesh))
			Expect(err).ToNot(HaveOccurred())

			name := "resource-1"
			trRes := newMeshExternalService()
			err = s.Create(context.Background(), trRes,
				store.CreateByKey(name, mesh),
				store.CreatedAt(time.Now()),
				store.CreateWithOwner(meshRes))
			Expect(err).ToNot(HaveOccurred())

			// when owner is updated
			Expect(s.Update(context.Background(), meshRes)).To(Succeed())

			// and only then deleted
			err = s.Delete(context.Background(), meshRes, store.DeleteByKey(mesh, model.NoMesh))
			Expect(err).ToNot(HaveOccurred())

			// then
			actual := meshexternalservice_api.NewMeshExternalServiceResource()
			err = s.Get(context.Background(), actual, store.GetByKey(name, mesh))
			Expect(store.IsNotFound(err)).To(BeTrue())
		})

		It("should delete several resources when their owner is deleted", func() {
			// setup
			meshRes := core_mesh.NewMeshResource()
			err := s.Create(context.Background(), meshRes, store.CreateByKey(mesh, model.NoMesh))
			Expect(err).ToNot(HaveOccurred())

			for i := range 10 {
				tr := newMeshExternalService()
				err = s.Create(context.Background(), tr,
					store.CreateByKey(fmt.Sprintf("resource-%d", i), mesh),
					store.CreatedAt(time.Now()),
					store.CreateWithOwner(meshRes))
				Expect(err).ToNot(HaveOccurred())
			}
			actual := meshexternalservice_api.MeshExternalServiceResourceList{}
			err = s.List(context.Background(), &actual, store.ListByMesh(mesh))
			Expect(err).ToNot(HaveOccurred())
			Expect(actual.Items).To(HaveLen(10))

			// when
			err = s.Delete(context.Background(), meshRes, store.DeleteByKey(mesh, model.NoMesh))
			Expect(err).ToNot(HaveOccurred())

			// then
			actual = meshexternalservice_api.MeshExternalServiceResourceList{}
			err = s.List(context.Background(), &actual, store.ListByMesh(mesh))
			Expect(err).ToNot(HaveOccurred())
			Expect(actual.Items).To(BeEmpty())
		})

		It("should delete owners chain", func() {
			// setup
			meshRes := core_mesh.NewMeshResource()
			err := s.Create(context.Background(), meshRes, store.CreateByKey(mesh, model.NoMesh))
			Expect(err).ToNot(HaveOccurred())

			var prev model.Resource = meshRes
			for i := range 10 {
				curr := newMeshExternalService()
				err := s.Create(context.Background(), curr,
					store.CreateByKey(fmt.Sprintf("resource-%d", i), mesh),
					store.CreatedAt(time.Now()),
					store.CreateWithOwner(prev))
				Expect(err).ToNot(HaveOccurred())
				prev = curr
			}

			actual := meshexternalservice_api.MeshExternalServiceResourceList{}
			err = s.List(context.Background(), &actual, store.ListByMesh(mesh))
			Expect(err).ToNot(HaveOccurred())
			Expect(actual.Items).To(HaveLen(10))

			// when
			err = s.Delete(context.Background(), meshRes, store.DeleteByKey(mesh, model.NoMesh))
			Expect(err).ToNot(HaveOccurred())

			// then
			actual = meshexternalservice_api.MeshExternalServiceResourceList{}
			err = s.List(context.Background(), &actual, store.ListByMesh(mesh))
			Expect(err).ToNot(HaveOccurred())
			Expect(actual.Items).To(BeEmpty())
		})

		It("should delete a parent after children is deleted", func() {
			// given
			meshRes := core_mesh.NewMeshResource()
			err := s.Create(context.Background(), meshRes, store.CreateByKey(mesh, model.NoMesh))
			Expect(err).ToNot(HaveOccurred())

			name := "resource-1"
			trRes := newMeshExternalService()
			err = s.Create(context.Background(), trRes,
				store.CreateByKey(name, mesh),
				store.CreatedAt(time.Now()),
				store.CreateWithOwner(meshRes))
			Expect(err).ToNot(HaveOccurred())

			// when children is deleted
			err = s.Delete(context.Background(), meshexternalservice_api.NewMeshExternalServiceResource(), store.DeleteByKey(name, mesh))

			// then
			Expect(err).ToNot(HaveOccurred())

			// when parent is deleted
			err = s.Delete(context.Background(), core_mesh.NewMeshResource(), store.DeleteByKey(mesh, model.NoMesh))

			// then
			Expect(err).ToNot(HaveOccurred())
		})
	})
}
