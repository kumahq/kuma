package v2

import (
	"fmt"
	"sync/atomic"

	envoy "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	envoy_auth "github.com/envoyproxy/go-control-plane/envoy/api/v2/auth"
	envoy_types "github.com/envoyproxy/go-control-plane/pkg/cache/types"
	envoy_cache "github.com/envoyproxy/go-control-plane/pkg/cache/v2"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core"
	mesh_core "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	xds_model "github.com/kumahq/kuma/pkg/core/xds"
	core_xds_v2 "github.com/kumahq/kuma/pkg/core/xds/v2"
	test_model "github.com/kumahq/kuma/pkg/test/resources/model"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
	envoy_common "github.com/kumahq/kuma/pkg/xds/envoy"
)

var _ = Describe("Reconcile", func() {
	Describe("reconciler", func() {

		var backupNewUUID func() string

		BeforeEach(func() {
			backupNewUUID = core.NewUUID
		})
		AfterEach(func() {
			core.NewUUID = backupNewUUID
		})

		var serial uint64

		BeforeEach(func() {
			core.NewUUID = func() string {
				uuid := atomic.AddUint64(&serial, 1)
				return fmt.Sprintf("v%d", uuid)
			}
		})

		var xdsContext core_xds_v2.XdsContext

		BeforeEach(func() {
			xdsContext = core_xds_v2.NewXdsContext()
		})

		snapshot := envoy_cache.Snapshot{
			Resources: [envoy_types.UnknownType]envoy_cache.Resources{
				envoy_types.Listener: {
					Items: map[string]envoy_types.Resource{
						"listener": &envoy.Listener{},
					},
				},
				envoy_types.Route: {
					Items: map[string]envoy_types.Resource{
						"route": &envoy.RouteConfiguration{},
					},
				},
				envoy_types.Cluster: {
					Items: map[string]envoy_types.Resource{
						"cluster": &envoy.Cluster{},
					},
				},
				envoy_types.Endpoint: {
					Items: map[string]envoy_types.Resource{
						"endpoint": &envoy.ClusterLoadAssignment{},
					},
				},
				envoy_types.Secret: {
					Items: map[string]envoy_types.Resource{
						"secret": &envoy_auth.Secret{},
					},
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
				Spec: &mesh_proto.Dataplane{},
			}

			By("simulating discovery event")
			// when
			proxy := &xds_model.Proxy{
				Id: xds_model.ProxyId{
					Mesh: "demo",
					Name: "example",
				},
				Dataplane:  dataplane,
				APIVersion: envoy_common.APIV2,
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
			Expect(snapshot.Resources[envoy_types.Listener].Version).To(Equal("v1"))
			Expect(snapshot.Resources[envoy_types.Route].Version).To(Equal("v2"))
			Expect(snapshot.Resources[envoy_types.Cluster].Version).To(Equal("v3"))
			Expect(snapshot.Resources[envoy_types.Endpoint].Version).To(Equal("v4"))
			Expect(snapshot.Resources[envoy_types.Secret].Version).To(Equal("v5"))

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
			Expect(snapshot.Resources[envoy_types.Listener].Version).To(Equal("v1"))
			Expect(snapshot.Resources[envoy_types.Route].Version).To(Equal("v2"))
			Expect(snapshot.Resources[envoy_types.Cluster].Version).To(Equal("v3"))
			Expect(snapshot.Resources[envoy_types.Endpoint].Version).To(Equal("v4"))
			Expect(snapshot.Resources[envoy_types.Secret].Version).To(Equal("v5"))

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
			Expect(snapshot.Resources[envoy_types.Listener].Version).To(Equal("v6"))
			Expect(snapshot.Resources[envoy_types.Route].Version).To(Equal("v7"))
			Expect(snapshot.Resources[envoy_types.Cluster].Version).To(Equal("v8"))
			Expect(snapshot.Resources[envoy_types.Endpoint].Version).To(Equal("v9"))
			Expect(snapshot.Resources[envoy_types.Secret].Version).To(Equal("v10"))
		})
	})
})

type snapshotGeneratorFunc func(ctx xds_context.Context, proxy *xds_model.Proxy) (envoy_cache.Snapshot, error)

func (f snapshotGeneratorFunc) GenerateSnapshot(ctx xds_context.Context, proxy *xds_model.Proxy) (envoy_cache.Snapshot, error) {
	return f(ctx, proxy)
}
