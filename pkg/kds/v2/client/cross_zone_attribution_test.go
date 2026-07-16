// KDS zone-to-global sync attribution guard.
//
// The global CP must attribute a synced resource by the authenticated
// connection identity (the client-id from util.ClientIDFromIncomingCtx, wired
// into the stream in pkg/kds/mux/zone_sync.go), not by values the sender
// provides in-band. These tests drive the real global ingest wiring
// (NewDeltaKDSStream(stream, <authenticated zone>, ...) -> GlobalSyncCallback ->
// KDSSyncClient.Receive) with a stream whose authenticated client-id ("zone-b")
// differs from the in-band ControlPlane.Identifier and the in-spec kuma.io/zone
// tags ("zone-a"), and assert that all attribution resolves to the authenticated
// client-id.
package client_test

import (
	"context"
	"fmt"
	"testing"

	envoy_core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	envoy_sd "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"
	. "github.com/onsi/gomega"
	"google.golang.org/grpc/metadata"

	mesh_proto "github.com/kumahq/kuma/v3/api/mesh/v1alpha1"
	system_proto "github.com/kumahq/kuma/v3/api/system/v1alpha1"
	"github.com/kumahq/kuma/v3/pkg/core"
	core_mesh "github.com/kumahq/kuma/v3/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/v3/pkg/core/resources/apis/system"
	core_model "github.com/kumahq/kuma/v3/pkg/core/resources/model"
	"github.com/kumahq/kuma/v3/pkg/core/resources/store"
	"github.com/kumahq/kuma/v3/pkg/kds"
	client_v2 "github.com/kumahq/kuma/v3/pkg/kds/v2/client"
	sync_store_v2 "github.com/kumahq/kuma/v3/pkg/kds/v2/store"
	core_metrics "github.com/kumahq/kuma/v3/pkg/metrics"
	"github.com/kumahq/kuma/v3/pkg/plugins/resources/memory"
	test_grpc "github.com/kumahq/kuma/v3/pkg/test/grpc"
	util_proto "github.com/kumahq/kuma/v3/pkg/util/proto"
)

const (
	// authenticatedZone is the connection's authenticated identity (client-id),
	// which zone_sync.go passes to NewDeltaKDSStream and which must drive
	// attribution.
	authenticatedZone = "zone-b"
	// otherZone is a different, already-enrolled zone named by the sender's
	// in-band ControlPlane.Identifier and in-spec kuma.io/zone tags.
	otherZone = "zone-a"
)

// newGlobalSink wires the global-CP KDS zone-to-global ingest as
// pkg/kds/mux/zone_sync.go does, feeding a memory-backed store. authClientID is
// the authenticated zone (client-id), intentionally different from the
// per-message ControlPlane.Identifier and in-spec tags the responses carry.
func newGlobalSink(t *testing.T, ctx context.Context, authClientID string, typ core_model.ResourceType) (store.ResourceStore, *test_grpc.MockDeltaClientStream, func()) {
	t.Helper()
	g := NewWithT(t)

	globalStore := memory.NewStore()
	// Both zones exist on global, as in any federated mesh.
	for _, z := range []string{otherZone, authenticatedZone} {
		err := globalStore.Create(ctx,
			&system.ZoneResource{Spec: &system_proto.Zone{Enabled: util_proto.Bool(true)}},
			store.CreateByKey(z, core_model.NoMesh))
		g.Expect(err).ToNot(HaveOccurred())
	}

	metrics, err := core_metrics.NewMetrics("")
	g.Expect(err).ToNot(HaveOccurred())
	syncer, err := sync_store_v2.NewResourceSyncer(core.Log.WithName("crosszone-syncer"), globalStore, store.NoTransactions{}, metrics, context.Background())
	g.Expect(err).ToNot(HaveOccurred())

	clientStream := test_grpc.NewMockDeltaClientStream()
	// authClientID is the authenticated zone that must drive attribution, not
	// the in-band ControlPlane.Identifier.
	kdsStream := client_v2.NewDeltaKDSStream(clientStream, authClientID, authClientID+"-instance", "", 1)
	sink := client_v2.NewKDSSyncClient(
		core.Log.WithName("crosszone-global-sink"),
		[]core_model.ResourceType{typ},
		kdsStream,
		sync_store_v2.GlobalSyncCallback(ctx, syncer, false, nil, "kuma-system"),
		client_v2.SyncClientConfig{},
	)

	done := make(chan struct{})
	go func() {
		defer close(done)
		_ = sink.Receive()
	}()

	cleanup := func() {
		close(clientStream.RecvCh)
		_ = kdsStream.CloseSend()
		<-done
	}
	return globalStore, clientStream, cleanup
}

// deltaResponse builds a KDS DeltaDiscoveryResponse carrying one resource and a
// chosen ControlPlane.Identifier - the wire shape a zone CP sends.
func deltaResponse(t *testing.T, typ core_model.ResourceType, controlPlaneID, name, mesh string, spec core_model.ResourceSpec) *envoy_sd.DeltaDiscoveryResponse {
	t.Helper()
	g := NewWithT(t)

	specAny, err := core_model.ToAny(spec)
	g.Expect(err).ToNot(HaveOccurred())
	kr := &mesh_proto.KumaResource{
		Meta: &mesh_proto.KumaResource_Meta{Name: name, Mesh: mesh},
		Spec: specAny,
	}
	krAny, err := util_proto.MarshalAnyDeterministic(kr)
	g.Expect(err).ToNot(HaveOccurred())

	return &envoy_sd.DeltaDiscoveryResponse{
		TypeUrl:      string(typ),
		Nonce:        "nonce-1",
		ControlPlane: &envoy_core.ControlPlane{Identifier: controlPlaneID},
		Resources: []*envoy_sd.Resource{{
			Name:     fmt.Sprintf("%s.%s", name, mesh),
			Version:  "v1",
			Resource: krAny,
		}},
	}
}

// dataplaneWithDivergentZoneTags returns a Dataplane whose inbound kuma.io/zone
// tag names a zone different from the connection's authenticated zone. That tag
// feeds the global service inventory via createOrUpdateServiceInsight for
// non-Exclusive meshes, so global ingest must resolve it to the authenticated
// zone.
func dataplaneWithDivergentZoneTags() *mesh_proto.Dataplane {
	return &mesh_proto.Dataplane{
		Networking: &mesh_proto.Dataplane_Networking{
			Address: "192.168.0.1",
			Inbound: []*mesh_proto.Dataplane_Networking_Inbound{{
				Port: 1212,
				Tags: map[string]string{
					mesh_proto.ZoneTag:    otherZone,
					mesh_proto.ServiceTag: "svc-a",
				},
			}},
			Outbound: []*mesh_proto.Dataplane_Networking_Outbound{{
				Port: 10000,
				Tags: map[string]string{
					mesh_proto.ServiceTag:  "web",
					mesh_proto.ProtocolTag: "http",
				},
			}},
		},
	}
}

// gatewayDataplaneWithDivergentZoneTag returns a gateway Dataplane whose gateway
// kuma.io/zone tag names a zone different from the connection's authenticated
// zone.
func gatewayDataplaneWithDivergentZoneTag() *mesh_proto.Dataplane {
	return &mesh_proto.Dataplane{
		Networking: &mesh_proto.Dataplane_Networking{
			Address: "192.168.0.2",
			Gateway: &mesh_proto.Dataplane_Networking_Gateway{
				Type: mesh_proto.Dataplane_Networking_Gateway_BUILTIN,
				Tags: map[string]string{
					mesh_proto.ZoneTag:    otherZone,
					mesh_proto.ServiceTag: "gateway-a",
				},
			},
		},
	}
}

// zoneIngressWithDivergentZoneTags returns a ZoneIngress whose AvailableServices
// carry a kuma.io/zone tag naming a zone different from the connection's
// authenticated zone. That tag drives cross-zone endpoint locality in
// fillIngressOutbounds, so global ingest must resolve it to the authenticated
// zone.
func zoneIngressWithDivergentZoneTags() *mesh_proto.ZoneIngress {
	return &mesh_proto.ZoneIngress{
		Zone: authenticatedZone,
		Networking: &mesh_proto.ZoneIngress_Networking{
			Address: "10.0.0.1",
			Port:    10001,
		},
		AvailableServices: []*mesh_proto.ZoneIngress_AvailableService{{
			Tags: map[string]string{
				mesh_proto.ServiceTag: "svc-a",
				mesh_proto.ZoneTag:    otherZone,
			},
		}},
	}
}

// hashSuffixCtx advertises the hash-suffix feature, as every currently
// supported zone does, so the modern (non-prefixing) global ingest path runs.
func hashSuffixCtx() context.Context {
	return metadata.NewIncomingContext(context.Background(), metadata.Pairs(kds.FeaturesMetadataKey, kds.FeatureHashSuffix))
}

func storedDataplane(t *testing.T, s store.ResourceStore, ctx context.Context) *core_mesh.DataplaneResource {
	t.Helper()
	g := NewWithT(t)
	var stored *core_mesh.DataplaneResource
	g.Eventually(func(g Gomega) {
		list := core_mesh.DataplaneResourceList{}
		g.Expect(s.List(ctx, &list)).To(Succeed())
		g.Expect(list.Items).To(HaveLen(1))
		stored = list.Items[0]
	}, "5s", "50ms").Should(Succeed())
	return stored
}

func storedZoneIngress(t *testing.T, s store.ResourceStore, ctx context.Context) *core_mesh.ZoneIngressResource {
	t.Helper()
	g := NewWithT(t)
	var stored *core_mesh.ZoneIngressResource
	g.Eventually(func(g Gomega) {
		list := core_mesh.ZoneIngressResourceList{}
		g.Expect(s.List(ctx, &list)).To(Succeed())
		g.Expect(list.Items).To(HaveLen(1))
		stored = list.Items[0]
	}, "5s", "50ms").Should(Succeed())
	return stored
}

// TestZoneToGlobalSyncAttribution asserts that every zone attribution the global
// CP derives from a synced resource follows the authenticated client-id, not the
// sender-provided in-band values. It covers the top-level metadata (kuma.io/zone
// label, ZoneIngress.Spec.Zone) and the zone tags carried inside the specs
// (ZoneIngress AvailableServices; Dataplane inbound/gateway), which are consumed
// downstream for endpoint locality and the global service inventory. The stream
// authenticates as authenticatedZone while the payload names otherZone
// throughout; all of it must resolve to authenticatedZone.
func TestZoneToGlobalSyncAttribution(t *testing.T) {
	t.Run("Dataplane", func(t *testing.T) {
		g := NewWithT(t)
		ctx := hashSuffixCtx()
		globalStore, clientStream, cleanup := newGlobalSink(t, ctx, authenticatedZone, core_mesh.DataplaneType)
		defer cleanup()

		clientStream.RecvCh <- deltaResponse(t, core_mesh.DataplaneType, otherZone, "dp-1", "default", dataplaneWithDivergentZoneTags())

		stored := storedDataplane(t, globalStore, ctx)
		zone := stored.GetMeta().GetLabels()[mesh_proto.ZoneTag]
		g.Expect(zone).To(Equal(authenticatedZone),
			"kuma.io/zone label must be the authenticated client-id %q, not the sender-provided zone %q", authenticatedZone, zone)

		inboundZone := stored.Spec.GetNetworking().GetInbound()[0].GetTags()[mesh_proto.ZoneTag]
		g.Expect(inboundZone).To(Equal(authenticatedZone),
			"Dataplane inbound kuma.io/zone tag must resolve to the authenticated client-id %q, not the sender-provided zone %q", authenticatedZone, inboundZone)
	})

	t.Run("DataplaneGateway", func(t *testing.T) {
		g := NewWithT(t)
		ctx := hashSuffixCtx()
		globalStore, clientStream, cleanup := newGlobalSink(t, ctx, authenticatedZone, core_mesh.DataplaneType)
		defer cleanup()

		clientStream.RecvCh <- deltaResponse(t, core_mesh.DataplaneType, otherZone, "gw-dp-1", "default", gatewayDataplaneWithDivergentZoneTag())

		stored := storedDataplane(t, globalStore, ctx)
		gatewayZone := stored.Spec.GetNetworking().GetGateway().GetTags()[mesh_proto.ZoneTag]
		g.Expect(gatewayZone).To(Equal(authenticatedZone),
			"Dataplane gateway kuma.io/zone tag must resolve to the authenticated client-id %q, not the sender-provided zone %q", authenticatedZone, gatewayZone)
	})

	t.Run("ZoneIngress", func(t *testing.T) {
		g := NewWithT(t)
		ctx := hashSuffixCtx()
		globalStore, clientStream, cleanup := newGlobalSink(t, ctx, authenticatedZone, core_mesh.ZoneIngressType)
		defer cleanup()

		clientStream.RecvCh <- deltaResponse(t, core_mesh.ZoneIngressType, otherZone, "zi-1", "", zoneIngressWithDivergentZoneTags())

		stored := storedZoneIngress(t, globalStore, ctx)
		g.Expect(stored.Spec.GetZone()).To(Equal(authenticatedZone),
			"ZoneIngress.Spec.Zone must be the authenticated client-id %q, not the sender-provided zone %q", authenticatedZone, stored.Spec.GetZone())
		g.Expect(stored.GetMeta().GetLabels()[mesh_proto.ZoneTag]).To(Equal(authenticatedZone))

		svcZone := stored.Spec.GetAvailableServices()[0].GetTags()[mesh_proto.ZoneTag]
		g.Expect(svcZone).To(Equal(authenticatedZone),
			"ZoneIngress AvailableServices kuma.io/zone tag must resolve to the authenticated client-id %q, not the sender-provided zone %q", authenticatedZone, svcZone)
	})
}
