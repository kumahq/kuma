package server

import (
	"fmt"
	"sync/atomic"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	mesh_core "github.com/Kong/kuma/pkg/core/resources/apis/mesh"
	core_xds "github.com/Kong/kuma/pkg/core/xds"
	xds_model "github.com/Kong/kuma/pkg/core/xds"
	test_model "github.com/Kong/kuma/pkg/test/resources/model"
	xds_context "github.com/Kong/kuma/pkg/xds/context"
	envoy "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	envoy_auth "github.com/envoyproxy/go-control-plane/envoy/api/v2/auth"
	envoy_cache "github.com/envoyproxy/go-control-plane/pkg/cache"
)

var _ = Describe("Reconcile", func() {
	Describe("reconciler", func() {

		var backupNewUUID func() string

		BeforeEach(func() {
			backupNewUUID = newUUID
		})
		AfterEach(func() {
			newUUID = backupNewUUID
		})

		var serial uint64

		BeforeEach(func() {
			newUUID = func() string {
				uuid := atomic.AddUint64(&serial, 1)
				return fmt.Sprintf("v%d", uuid)
			}
		})

		var xdsContext core_xds.XdsContext

		BeforeEach(func() {
			xdsContext = core_xds.NewXdsContext()
		})

		snapshot := envoy_cache.Snapshot{
			Listeners: envoy_cache.Resources{
				Items: map[string]envoy_cache.Resource{
					"listener": &envoy.Listener{},
				},
			},
			Routes: envoy_cache.Resources{
				Items: map[string]envoy_cache.Resource{
					"route": &envoy.RouteConfiguration{},
				},
			},
			Clusters: envoy_cache.Resources{
				Items: map[string]envoy_cache.Resource{
					"cluster": &envoy.Cluster{},
				},
			},
			Endpoints: envoy_cache.Resources{
				Items: map[string]envoy_cache.Resource{
					"endpoint": &envoy.ClusterLoadAssignment{},
				},
			},
			Secrets: envoy_cache.Resources{
				Items: map[string]envoy_cache.Resource{
					"secret": &envoy_auth.Secret{},
				},
			},
		}

		It("should generate a Snaphot per Envoy Node", func() {
			// given
			snapshots := make(chan envoy_cache.Snapshot, 3)
			snapshots <- snapshot               // initial Dataplane configuration
			snapshots <- snapshot               // same Dataplane configuration
			snapshots <- envoy_cache.Snapshot{} // new Dataplane configuration

			// setup
			r := &reconciler{
				snapshotGeneratorFunc(func(ctx xds_context.Context, proxy *xds_model.Proxy) (envoy_cache.Snapshot, error) {
					return <-snapshots, nil
				}),
				&simpleSnapshotCacher{xdsContext.Hasher(), xdsContext.Cache()},
			}

			// given
			dataplane := &mesh_core.DataplaneResource{
				Meta: &test_model.ResourceMeta{
					Mesh:    "demo",
					Name:    "example",
					Version: "abcdefg",
				},
			}

			By("simulating discovery event")
			// when
			proxy := &xds_model.Proxy{
				Id: xds_model.ProxyId{
					Mesh: "demo",
					Name: "example",
				},
				Dataplane: dataplane,
			}
			err := r.Reconcile(xds_context.Context{}, proxy)
			// then
			Expect(err).ToNot(HaveOccurred())

			By("verifying that snapshot versions were auto-generated")
			// when
			snapshot, err := xdsContext.Cache().GetSnapshot("demo.example")
			// then
			Expect(err).ToNot(HaveOccurred())
			Expect(snapshot).ToNot(BeZero())
			// and
			Expect(snapshot.Listeners.Version).To(Equal("v1"))
			Expect(snapshot.Routes.Version).To(Equal("v2"))
			Expect(snapshot.Clusters.Version).To(Equal("v3"))
			Expect(snapshot.Endpoints.Version).To(Equal("v4"))
			Expect(snapshot.Secrets.Version).To(Equal("v5"))

			By("simulating discovery event (Dataplane watchdog triggers refresh)")
			// when
			err = r.Reconcile(xds_context.Context{}, proxy)
			// then
			Expect(err).ToNot(HaveOccurred())

			By("verifying that snapshot versions remain the same")
			// when
			snapshot, err = xdsContext.Cache().GetSnapshot("demo.example")
			// then
			Expect(err).ToNot(HaveOccurred())
			Expect(snapshot).ToNot(BeZero())
			// and
			Expect(snapshot.Listeners.Version).To(Equal("v1"))
			Expect(snapshot.Routes.Version).To(Equal("v2"))
			Expect(snapshot.Clusters.Version).To(Equal("v3"))
			Expect(snapshot.Endpoints.Version).To(Equal("v4"))
			Expect(snapshot.Secrets.Version).To(Equal("v5"))

			By("simulating discovery event (Dataplane gets changed)")
			// when
			err = r.Reconcile(xds_context.Context{}, proxy)
			// then
			Expect(err).ToNot(HaveOccurred())

			By("verifying that snapshot versions are new")
			// when
			snapshot, err = xdsContext.Cache().GetSnapshot("demo.example")
			// then
			Expect(err).ToNot(HaveOccurred())
			Expect(snapshot).ToNot(BeZero())
			// and
			Expect(snapshot.Listeners.Version).To(Equal("v6"))
			Expect(snapshot.Routes.Version).To(Equal("v7"))
			Expect(snapshot.Clusters.Version).To(Equal("v8"))
			Expect(snapshot.Endpoints.Version).To(Equal("v9"))
			Expect(snapshot.Secrets.Version).To(Equal("v10"))
		})
	})
})

type snapshotGeneratorFunc func(ctx xds_context.Context, proxy *xds_model.Proxy) (envoy_cache.Snapshot, error)

func (f snapshotGeneratorFunc) GenerateSnapshot(ctx xds_context.Context, proxy *xds_model.Proxy) (envoy_cache.Snapshot, error) {
	return f(ctx, proxy)
}
