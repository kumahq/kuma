package manager_test

import (
	"context"
	"fmt"
	"sync/atomic"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	mesh_proto "github.com/kumahq/kuma/v2/api/mesh/v1alpha1"
	core_mesh "github.com/kumahq/kuma/v2/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/v2/pkg/core/resources/manager"
	"github.com/kumahq/kuma/v2/pkg/core/resources/model"
	"github.com/kumahq/kuma/v2/pkg/core/resources/store"
	"github.com/kumahq/kuma/v2/pkg/plugins/resources/memory"
)

type safeToRetryError struct {
	msg string
}

func (e *safeToRetryError) Error() string     { return e.msg }
func (e *safeToRetryError) SafeToRetry() bool { return true }
func (e *safeToRetryError) Unwrap() error     { return nil }

type unsafeError struct {
	msg string
}

func (e *unsafeError) Error() string     { return e.msg }
func (e *unsafeError) SafeToRetry() bool { return false }

// failingStore wraps a ResourceStore and fails Get() calls
// with the given error for the first N attempts.
type failingStore struct {
	store.ResourceStore
	failTimes int32
	attempts  atomic.Int32
	err       error
}

func (s *failingStore) Get(ctx context.Context, r model.Resource, fs ...store.GetOptionsFunc) error {
	if s.attempts.Add(1) <= s.failTimes {
		return s.err
	}
	return s.ResourceStore.Get(ctx, r, fs...)
}

var _ = Describe("Resource Manager", func() {
	var resStore store.ResourceStore
	var resManager manager.ResourceManager

	BeforeEach(func() {
		resStore = memory.NewStore()
		resManager = manager.NewResourceManager(resStore)
	})

	createSampleMesh := func(name string) error {
		meshRes := core_mesh.MeshResource{
			Spec: &mesh_proto.Mesh{},
		}
		return resManager.Create(context.Background(), &meshRes, store.CreateByKey(name, model.NoMesh))
	}

	createSampleResource := func(mesh string) (*core_mesh.TrafficRouteResource, error) {
		trRes := core_mesh.TrafficRouteResource{
			Spec: &mesh_proto.TrafficRoute{
				Sources: []*mesh_proto.Selector{{Match: map[string]string{
					mesh_proto.ServiceTag: "*",
				}}},
				Destinations: []*mesh_proto.Selector{{Match: map[string]string{
					mesh_proto.ServiceTag: "*",
				}}},
				Conf: &mesh_proto.TrafficRoute_Conf{
					Destination: map[string]string{
						mesh_proto.ServiceTag: "*",
						"path":                "demo",
					},
				},
			},
		}
		err := resManager.Create(context.Background(), &trRes, store.CreateByKey("tr-1", mesh))
		return &trRes, err
	}

	Describe("Create()", func() {
		It("should let create when mesh exists", func() {
			// given
			err := createSampleMesh("mesh-1")
			Expect(err).ToNot(HaveOccurred())

			// when
			_, err = createSampleResource("mesh-1")

			// then
			Expect(err).ToNot(HaveOccurred())
		})

		It("should not let to create a resource when mesh not exists", func() {
			// given no mesh for resource

			// when
			_, err := createSampleResource("mesh-1")

			// then
			Expect(err.Error()).To(Equal("mesh of name mesh-1 is not found"))
		})
	})

	Describe("DeleteAll()", func() {
		It("should delete all resources within a mesh", func() {
			// setup
			Expect(createSampleMesh("mesh-1")).To(Succeed())
			Expect(createSampleMesh("mesh-2")).To(Succeed())
			_, err := createSampleResource("mesh-1")
			Expect(err).ToNot(HaveOccurred())
			_, err = createSampleResource("mesh-2")
			Expect(err).ToNot(HaveOccurred())

			tlKey := model.ResourceKey{
				Mesh: "mesh-1",
				Name: "tl-1",
			}
			trafficLog := &core_mesh.TrafficLogResource{
				Spec: &mesh_proto.TrafficLog{
					Sources: []*mesh_proto.Selector{
						{
							Match: map[string]string{
								"kuma.io/service": "*",
							},
						},
					},
					Destinations: []*mesh_proto.Selector{
						{
							Match: map[string]string{
								"kuma.io/service": "*",
							},
						},
					},
				},
			}
			err = resManager.Create(context.Background(), trafficLog, store.CreateBy(tlKey))
			Expect(err).ToNot(HaveOccurred())

			// when
			err = resManager.DeleteAll(context.Background(), &core_mesh.TrafficRouteResourceList{}, store.DeleteAllByMesh("mesh-1"))

			// then
			Expect(err).ToNot(HaveOccurred())

			// and resource from mesh-1 is deleted
			res1 := core_mesh.NewTrafficRouteResource()
			err = resManager.Get(context.Background(), res1, store.GetByKey("tr-1", "mesh-1"))
			Expect(store.IsNotFound(err)).To(BeTrue())

			// and only TrafficRoutes are deleted
			Expect(resManager.Get(context.Background(), core_mesh.NewTrafficLogResource(), store.GetBy(tlKey))).To(Succeed())

			// and resource from mesh-2 is retained
			res2 := core_mesh.NewTrafficRouteResource()
			err = resManager.Get(context.Background(), res2, store.GetByKey("tr-1", "mesh-2"))
			Expect(err).ToNot(HaveOccurred())
		})
	})

	Describe("Upsert()", func() {
		It("should retry on SafeToRetry errors", func() {
			// given
			underlying := memory.NewStore()
			fStore := &failingStore{
				ResourceStore: underlying,
				failTimes:     2,
				err:           &safeToRetryError{msg: "conn closed"},
			}
			mgr := manager.NewResourceManager(fStore)

			meshRes := core_mesh.MeshResource{
				Spec: &mesh_proto.Mesh{},
			}
			Expect(mgr.Create(context.Background(), &meshRes,
				store.CreateByKey("default", model.NoMesh))).To(Succeed())

			// when
			res := core_mesh.NewTrafficRouteResource()
			key := model.ResourceKey{Mesh: "default", Name: "tr-1"}
			err := manager.Upsert(
				context.Background(), mgr, key, res,
				func(r model.Resource) error {
					tr := r.(*core_mesh.TrafficRouteResource)
					tr.Spec = &mesh_proto.TrafficRoute{
						Sources: []*mesh_proto.Selector{{
							Match: map[string]string{
								mesh_proto.ServiceTag: "*",
							},
						}},
						Destinations: []*mesh_proto.Selector{{
							Match: map[string]string{
								mesh_proto.ServiceTag: "*",
							},
						}},
						Conf: &mesh_proto.TrafficRoute_Conf{
							Destination: map[string]string{
								mesh_proto.ServiceTag: "backend",
							},
						},
					}
					return nil
				},
				manager.WithConflictRetry(
					1*time.Millisecond, 5, 1,
				),
			)

			// then — retried through failures and succeeded
			Expect(err).ToNot(HaveOccurred())
			Expect(fStore.attempts.Load()).To(
				BeNumerically(">=", 3),
				"expected at least 2 failed + 1 successful Get",
			)
		})

		It("should not retry when SafeToRetry returns false", func() {
			// given
			underlying := memory.NewStore()
			fStore := &failingStore{
				ResourceStore: underlying,
				failTimes:     2,
				err:           &unsafeError{msg: "fatal error"},
			}
			mgr := manager.NewResourceManager(fStore)

			meshRes := core_mesh.MeshResource{
				Spec: &mesh_proto.Mesh{},
			}
			Expect(mgr.Create(context.Background(), &meshRes,
				store.CreateByKey("default", model.NoMesh))).To(Succeed())

			// when
			res := core_mesh.NewTrafficRouteResource()
			key := model.ResourceKey{Mesh: "default", Name: "tr-1"}
			err := manager.Upsert(
				context.Background(), mgr, key, res,
				func(r model.Resource) error { return nil },
				manager.WithConflictRetry(
					1*time.Millisecond, 5, 1,
				),
			)

			// then — error returned immediately, no retry
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("fatal error"))
			Expect(fStore.attempts.Load()).To(
				BeNumerically("==", 1),
				fmt.Sprintf(
					"expected exactly 1 attempt, got %d",
					fStore.attempts.Load(),
				),
			)
		})
	})
})
