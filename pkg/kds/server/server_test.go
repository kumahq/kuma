package server_test

import (
	"context"
	"time"

	"github.com/gogo/protobuf/proto"

	system_proto "github.com/Kong/kuma/api/system/v1alpha1"
	"github.com/Kong/kuma/pkg/core/resources/apis/system"
	"github.com/Kong/kuma/pkg/kds"
	test_grpc "github.com/Kong/kuma/pkg/test/grpc"
	kds_verifier "github.com/Kong/kuma/pkg/test/kds/verifier"

	v2 "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	envoy_core "github.com/envoyproxy/go-control-plane/envoy/api/v2/core"
	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/duration"
	"github.com/golang/protobuf/ptypes/wrappers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	mesh_proto "github.com/Kong/kuma/api/mesh/v1alpha1"
	kuma_cp "github.com/Kong/kuma/pkg/config/app/kuma-cp"
	kds_config "github.com/Kong/kuma/pkg/config/kds"
	"github.com/Kong/kuma/pkg/core"
	"github.com/Kong/kuma/pkg/core/resources/apis/mesh"
	"github.com/Kong/kuma/pkg/core/resources/manager"
	"github.com/Kong/kuma/pkg/core/resources/store"
	"github.com/Kong/kuma/pkg/core/runtime"
	"github.com/Kong/kuma/pkg/core/runtime/component"
	kds_server "github.com/Kong/kuma/pkg/kds/server"
	"github.com/Kong/kuma/pkg/plugins/resources/memory"
	util_xds "github.com/Kong/kuma/pkg/util/xds"
)

var (
	node = &envoy_core.Node{
		Id:      "test-id",
		Cluster: "test-cluster",
	}
)

const (
	defaultTimeout = 3 * time.Second
)

type testRuntimeContext struct {
	runtime.Runtime
	rom        manager.ReadOnlyResourceManager
	cfg        kuma_cp.Config
	components []component.Component
}

func (t *testRuntimeContext) Config() kuma_cp.Config {
	return t.cfg
}

func (t *testRuntimeContext) ReadOnlyResourceManager() manager.ReadOnlyResourceManager {
	return t.rom
}

func (t *testRuntimeContext) Add(c ...component.Component) error {
	t.components = append(t.components, c...)
	return nil
}

var (
	mesh1 = mesh_proto.Mesh{
		Mtls: &mesh_proto.Mesh_Mtls{
			EnabledBackend: "ca-1",
			Backends: []*mesh_proto.CertificateAuthorityBackend{
				{
					Name: "ca-1",
					Type: "builtin",
				},
			},
		},
	}
	mesh2 = mesh_proto.Mesh{
		Mtls: &mesh_proto.Mesh_Mtls{
			EnabledBackend: "ca-2",
			Backends: []*mesh_proto.CertificateAuthorityBackend{
				{
					Name: "ca-2",
					Type: "builtin",
				},
			},
		},
	}
	fi1 = mesh_proto.FaultInjection{
		Sources: []*mesh_proto.Selector{{
			Match: map[string]string{
				"service": "*",
				"tag0":    "version0",
				"tag1":    "version1",
				"tag2":    "version2",
				"tag3":    "version3",
				"tag4":    "version4",
				"tag5":    "version5",
				"tag6":    "version6",
				"tag7":    "version7",
				"tag8":    "version8",
				"tag9":    "version9",
			},
		}},
		Destinations: []*mesh_proto.Selector{{
			Match: map[string]string{
				"service": "*",
			},
		}},
		Conf: &mesh_proto.FaultInjection_Conf{
			Abort: &mesh_proto.FaultInjection_Conf_Abort{
				Percentage: &wrappers.DoubleValue{Value: 90},
				HttpStatus: &wrappers.UInt32Value{Value: 404},
			},
		},
	}
	ingress = mesh_proto.Dataplane{
		Networking: &mesh_proto.Dataplane_Networking{
			Ingress: &mesh_proto.Dataplane_Networking_Ingress{
				AvailableServices: []*mesh_proto.Dataplane_Networking_Ingress_AvailableService{{
					Tags: map[string]string{
						"service": "backend",
					}},
				},
			},
			Address: "192.168.0.1",
		},
	}
	cb1 = mesh_proto.CircuitBreaker{
		Sources: []*mesh_proto.Selector{{
			Match: map[string]string{
				"service": "*",
			},
		}},
		Destinations: []*mesh_proto.Selector{{
			Match: map[string]string{
				"service": "*",
			},
		}},
		Conf: &mesh_proto.CircuitBreaker_Conf{
			Detectors: &mesh_proto.CircuitBreaker_Conf_Detectors{},
		},
	}
	hc1 = mesh_proto.HealthCheck{
		Sources: []*mesh_proto.Selector{{
			Match: map[string]string{
				"service": "*",
			},
		}},
		Destinations: []*mesh_proto.Selector{{
			Match: map[string]string{
				"service": "*",
			},
		}},
		Conf: &mesh_proto.HealthCheck_Conf{
			ActiveChecks: &mesh_proto.HealthCheck_Conf_Active{
				Interval: &duration.Duration{Seconds: 5},
				Timeout:  &duration.Duration{Seconds: 7},
			},
		},
	}
	tl1 = mesh_proto.TrafficLog{
		Sources: []*mesh_proto.Selector{{
			Match: map[string]string{
				"service": "*",
			},
		}},
		Destinations: []*mesh_proto.Selector{{
			Match: map[string]string{
				"service": "*",
			},
		}},
		Conf: &mesh_proto.TrafficLog_Conf{
			Backend: "logging-backend",
		},
	}
	tp1 = mesh_proto.TrafficPermission{
		Sources: []*mesh_proto.Selector{{
			Match: map[string]string{
				"service": "*",
			},
		}},
		Destinations: []*mesh_proto.Selector{{
			Match: map[string]string{
				"service": "*",
			},
		}},
	}
	tr1 = mesh_proto.TrafficRoute{
		Sources: []*mesh_proto.Selector{{
			Match: map[string]string{
				"service": "*",
			},
		}},
		Destinations: []*mesh_proto.Selector{{
			Match: map[string]string{
				"service": "*",
			},
		}},
		Conf: []*mesh_proto.TrafficRoute_WeightedDestination{{
			Weight: 10,
			Destination: map[string]string{
				"version": "v2",
			},
		}},
	}
	tt1 = mesh_proto.TrafficTrace{
		Selectors: []*mesh_proto.Selector{{
			Match: map[string]string{"serivce": "*"},
		}},
		Conf: &mesh_proto.TrafficTrace_Conf{
			Backend: "tracing-backend",
		},
	}
	pt1 = mesh_proto.ProxyTemplate{
		Selectors: []*mesh_proto.Selector{{
			Match: map[string]string{"serivce": "*"},
		}},
		Conf: &mesh_proto.ProxyTemplate_Conf{
			Imports: []string{"default-kuma-profile"},
		},
	}
	s1 = system_proto.Secret{
		Data: &wrappers.BytesValue{Value: []byte("secret key")},
	}
)

var _ = Describe("KDS Server", func() {
	createServer := func(rt runtime.Runtime) kds_server.Server {
		log := core.Log
		hasher, cache := kds_server.NewXdsContext(log)
		generator := kds_server.NewSnapshotGenerator(rt)
		versioner := kds_server.NewVersioner()
		reconciler := kds_server.NewReconciler(hasher, cache, generator, versioner)
		syncTracker := kds_server.NewSyncTracker(reconciler, rt.Config().KDSServer.RefreshInterval)
		callbacks := util_xds.CallbacksChain{
			util_xds.LoggingCallbacks{Log: log},
			syncTracker,
		}
		return kds_server.NewServer(cache, callbacks, log)
	}

	var tc kds_verifier.TestContext
	BeforeEach(func() {
		s := memory.NewStore()
		mgr := manager.NewResourceManager(s)
		srv := createServer(&testRuntimeContext{
			rom: mgr,
			cfg: kuma_cp.Config{
				KDSServer: &kds_config.KumaDiscoveryServerConfig{
					RefreshInterval: 100 * time.Millisecond,
				},
			},
		})
		stream := test_grpc.MakeMockStream()
		stop := make(chan struct{})
		go func() {
			err := srv.StreamKumaResources(stream)
			Expect(err).ToNot(HaveOccurred())
			close(stop)
		}()

		tc = &kds_verifier.TestContextImpl{
			ResourceStore: s,
			MockStream:    stream,
			StopCh:        stop,
			Responses:     map[string]*v2.DiscoveryResponse{},
		}
	})

	It("should support all existing resource types", func() {
		ctx := context.Background()

		// Just to don't forget to update this test after updating 'kds.SupportedTypes
		Expect([]proto.Message{&mesh1, &ingress, &cb1, &fi1, &hc1, &tl1, &tp1, &tr1, &tt1, &pt1, &s1}).To(HaveLen(len(kds.SupportedTypes)))

		vrf := kds_verifier.New().
			Exec(kds_verifier.Create(ctx, &mesh.MeshResource{Spec: mesh1}, store.CreateByKey("mesh-1", "mesh-1"))).
			Exec(kds_verifier.Create(ctx, &mesh.DataplaneResource{Spec: ingress}, store.CreateByKey("ingress-1", "mesh-1"))).
			Exec(kds_verifier.Create(ctx, &mesh.CircuitBreakerResource{Spec: cb1}, store.CreateByKey("cb-1", "mesh-1"))).
			Exec(kds_verifier.Create(ctx, &mesh.FaultInjectionResource{Spec: fi1}, store.CreateByKey("fi-1", "mesh-1"))).
			Exec(kds_verifier.Create(ctx, &mesh.HealthCheckResource{Spec: hc1}, store.CreateByKey("hc-1", "mesh-1"))).
			Exec(kds_verifier.Create(ctx, &mesh.TrafficLogResource{Spec: tl1}, store.CreateByKey("tl-1", "mesh-1"))).
			Exec(kds_verifier.Create(ctx, &mesh.TrafficPermissionResource{Spec: tp1}, store.CreateByKey("tp-1", "mesh-1"))).
			Exec(kds_verifier.Create(ctx, &mesh.TrafficRouteResource{Spec: tr1}, store.CreateByKey("tr-1", "mesh-1"))).
			Exec(kds_verifier.Create(ctx, &mesh.TrafficTraceResource{Spec: tt1}, store.CreateByKey("tt-1", "mesh-1"))).
			Exec(kds_verifier.Create(ctx, &mesh.ProxyTemplateResource{Spec: pt1}, store.CreateByKey("pt-1", "mesh-1"))).
			Exec(kds_verifier.Create(ctx, &system.SecretResource{Spec: s1}, store.CreateByKey("s-1", "mesh-1"))).
			Exec(kds_verifier.DiscoveryRequest(node, mesh.MeshType)).
			Exec(kds_verifier.WaitResponse(defaultTimeout, func(krs []*mesh_proto.KumaResource) {
				Expect(krs).To(HaveLen(1))
				m := &mesh_proto.Mesh{}
				Expect(ptypes.UnmarshalAny(krs[0].Spec, m)).ToNot(HaveOccurred())
				Expect(m).To(Equal(&mesh1))
			})).
			Exec(kds_verifier.DiscoveryRequest(node, mesh.DataplaneType)).
			Exec(kds_verifier.WaitResponse(defaultTimeout, func(krs []*mesh_proto.KumaResource) {
				Expect(krs).To(HaveLen(1))
				m := &mesh_proto.Dataplane{}
				Expect(ptypes.UnmarshalAny(krs[0].Spec, m)).ToNot(HaveOccurred())
				Expect(m).To(Equal(&ingress))
			})).
			Exec(kds_verifier.DiscoveryRequest(node, mesh.CircuitBreakerType)).
			Exec(kds_verifier.WaitResponse(defaultTimeout, func(krs []*mesh_proto.KumaResource) {
				Expect(krs).To(HaveLen(1))
				m := &mesh_proto.CircuitBreaker{}
				Expect(ptypes.UnmarshalAny(krs[0].Spec, m)).ToNot(HaveOccurred())
				Expect(m).To(Equal(&cb1))
			})).
			Exec(kds_verifier.DiscoveryRequest(node, mesh.FaultInjectionType)).
			Exec(kds_verifier.WaitResponse(defaultTimeout, func(krs []*mesh_proto.KumaResource) {
				Expect(krs).To(HaveLen(1))
				m := &mesh_proto.FaultInjection{}
				Expect(ptypes.UnmarshalAny(krs[0].Spec, m)).ToNot(HaveOccurred())
				Expect(m).To(Equal(&fi1))
			})).
			Exec(kds_verifier.DiscoveryRequest(node, mesh.HealthCheckType)).
			Exec(kds_verifier.WaitResponse(defaultTimeout, func(krs []*mesh_proto.KumaResource) {
				Expect(krs).To(HaveLen(1))
				m := &mesh_proto.HealthCheck{}
				Expect(ptypes.UnmarshalAny(krs[0].Spec, m)).ToNot(HaveOccurred())
				Expect(m).To(Equal(&hc1))
			})).
			Exec(kds_verifier.DiscoveryRequest(node, mesh.TrafficLogType)).
			Exec(kds_verifier.WaitResponse(defaultTimeout, func(krs []*mesh_proto.KumaResource) {
				Expect(krs).To(HaveLen(1))
				m := &mesh_proto.TrafficLog{}
				Expect(ptypes.UnmarshalAny(krs[0].Spec, m)).ToNot(HaveOccurred())
				Expect(m).To(Equal(&tl1))
			})).
			Exec(kds_verifier.DiscoveryRequest(node, mesh.TrafficPermissionType)).
			Exec(kds_verifier.WaitResponse(defaultTimeout, func(krs []*mesh_proto.KumaResource) {
				Expect(krs).To(HaveLen(1))
				m := &mesh_proto.TrafficPermission{}
				Expect(ptypes.UnmarshalAny(krs[0].Spec, m)).ToNot(HaveOccurred())
				Expect(m).To(Equal(&tp1))
			})).
			Exec(kds_verifier.DiscoveryRequest(node, mesh.TrafficRouteType)).
			Exec(kds_verifier.WaitResponse(defaultTimeout, func(krs []*mesh_proto.KumaResource) {
				Expect(krs).To(HaveLen(1))
				m := &mesh_proto.TrafficRoute{}
				Expect(ptypes.UnmarshalAny(krs[0].Spec, m)).ToNot(HaveOccurred())
				Expect(m).To(Equal(&tr1))
			})).
			Exec(kds_verifier.DiscoveryRequest(node, mesh.TrafficTraceType)).
			Exec(kds_verifier.WaitResponse(defaultTimeout, func(krs []*mesh_proto.KumaResource) {
				Expect(krs).To(HaveLen(1))
				m := &mesh_proto.TrafficTrace{}
				Expect(ptypes.UnmarshalAny(krs[0].Spec, m)).ToNot(HaveOccurred())
				Expect(m).To(Equal(&tt1))
			})).
			Exec(kds_verifier.DiscoveryRequest(node, mesh.ProxyTemplateType)).
			Exec(kds_verifier.WaitResponse(defaultTimeout, func(krs []*mesh_proto.KumaResource) {
				Expect(krs).To(HaveLen(1))
				m := &mesh_proto.ProxyTemplate{}
				Expect(ptypes.UnmarshalAny(krs[0].Spec, m)).ToNot(HaveOccurred())
				Expect(m).To(Equal(&pt1))
			})).
			Exec(kds_verifier.DiscoveryRequest(node, system.SecretType)).
			Exec(kds_verifier.WaitResponse(defaultTimeout, func(krs []*mesh_proto.KumaResource) {
				Expect(krs).To(HaveLen(1))
				m := &system_proto.Secret{}
				Expect(ptypes.UnmarshalAny(krs[0].Spec, m)).ToNot(HaveOccurred())
				Expect(m).To(Equal(&s1))
			})).
			Exec(kds_verifier.CloseStream())

		err := vrf.Verify(tc)
		Expect(err).ToNot(HaveOccurred())

		<-tc.Stop()
	})

	It("should accept request independently for each type", func() {
		ctx := context.Background()

		vrf := kds_verifier.New().
			Exec(kds_verifier.Create(ctx, &mesh.MeshResource{Spec: mesh1}, store.CreateByKey("mesh-1", "mesh-1"))).
			Exec(kds_verifier.Create(ctx, &mesh.FaultInjectionResource{Spec: fi1}, store.CreateByKey("fi1", "mesh-2"))).
			Exec(kds_verifier.DiscoveryRequest(node, mesh.MeshType)).
			Exec(kds_verifier.DiscoveryRequest(node, mesh.FaultInjectionType)).
			Exec(kds_verifier.WaitResponse(defaultTimeout, func(krs []*mesh_proto.KumaResource) {
				Expect(krs).To(HaveLen(1))
			})).
			Exec(kds_verifier.WaitResponse(defaultTimeout, func(krs []*mesh_proto.KumaResource) {
				Expect(krs).To(HaveLen(1))
			})).
			Exec(kds_verifier.Ack(node, mesh.MeshType)).
			Exec(kds_verifier.Ack(node, mesh.FaultInjectionType)).
			Exec(kds_verifier.CloseStream())

		err := vrf.Verify(tc)
		Expect(err).ToNot(HaveOccurred())

		<-tc.Stop()
	})

	It("should send response for resources created after DiscoveryRequest", func() {
		ctx := context.Background()

		vrf := kds_verifier.New().
			Exec(kds_verifier.DiscoveryRequest(node, mesh.MeshType)).
			Exec(kds_verifier.Create(ctx, &mesh.MeshResource{Spec: mesh1}, store.CreateByKey("mesh-1", "mesh-1"))).
			Exec(kds_verifier.WaitResponse(defaultTimeout, func(krs []*mesh_proto.KumaResource) {
				Expect(krs).To(HaveLen(1))
			})).
			Exec(kds_verifier.Ack(node, mesh.MeshType)).
			Exec(kds_verifier.CloseStream())

		err := vrf.Verify(tc)
		Expect(err).ToNot(HaveOccurred())

		<-tc.Stop()
	})

	It("should send response for resources created before ACK", func() {
		ctx := context.Background()

		vrf := kds_verifier.New().
			Exec(kds_verifier.Create(ctx, &mesh.MeshResource{Spec: mesh1}, store.CreateByKey("mesh-1", "mesh-1"))).
			Exec(kds_verifier.DiscoveryRequest(node, mesh.MeshType)).
			Exec(kds_verifier.WaitResponse(defaultTimeout, func(krs []*mesh_proto.KumaResource) {
				Expect(krs).To(HaveLen(1))
			})).
			Exec(kds_verifier.Create(ctx, &mesh.MeshResource{Spec: mesh2}, store.CreateByKey("mesh-2", "mesh-2"))).
			Exec(kds_verifier.Ack(node, mesh.MeshType)).
			Exec(kds_verifier.WaitResponse(10*time.Second, func(krs []*mesh_proto.KumaResource) {
				Expect(krs).To(HaveLen(2))
			})).
			Exec(kds_verifier.CloseStream())

		err := vrf.Verify(tc)
		Expect(err).ToNot(HaveOccurred())

		<-tc.Stop()
	})

	It("should support update", func() {
		ctx := context.Background()

		vrf := kds_verifier.New().
			Exec(kds_verifier.Create(ctx, &mesh.MeshResource{Spec: mesh1}, store.CreateByKey("mesh-1", "mesh-1"))).
			Exec(kds_verifier.DiscoveryRequest(node, mesh.MeshType)).
			Exec(kds_verifier.WaitResponse(defaultTimeout, func(krs []*mesh_proto.KumaResource) {
				Expect(krs).To(HaveLen(1))
				m := &mesh_proto.Mesh{}
				Expect(ptypes.UnmarshalAny(krs[0].Spec, m)).ToNot(HaveOccurred())
				Expect(m).To(Equal(&mesh1))
			})).
			Exec(kds_verifier.Ack(node, mesh.MeshType)).
			Exec(func(tc kds_verifier.TestContext) error {
				var meshRes mesh.MeshResource
				if err := tc.Store().Get(ctx, &meshRes, store.GetByKey("mesh-1", "mesh-1")); err != nil {
					return err
				}
				meshRes.Spec = mesh2
				if err := tc.Store().Update(ctx, &meshRes); err != nil {
					return err
				}
				return nil
			}).
			Exec(kds_verifier.WaitResponse(defaultTimeout, func(krs []*mesh_proto.KumaResource) {
				Expect(krs).To(HaveLen(1))
				m := &mesh_proto.Mesh{}
				Expect(ptypes.UnmarshalAny(krs[0].Spec, m)).ToNot(HaveOccurred())
				Expect(m).To(Equal(&mesh2))
			})).
			Exec(kds_verifier.CloseStream())

		err := vrf.Verify(tc)
		Expect(err).ToNot(HaveOccurred())

		<-tc.Stop()
	})

	It("should have deterministic MarshalAny to avoid excess snapshot versions", func() {
		ctx := context.Background()

		vrf := kds_verifier.New().
			Exec(kds_verifier.Create(ctx, &mesh.FaultInjectionResource{Spec: fi1}, store.CreateByKey("fi-1", "mesh-1"))).
			Exec(kds_verifier.DiscoveryRequest(node, mesh.FaultInjectionType)).
			Exec(kds_verifier.WaitResponse(defaultTimeout, func(krs []*mesh_proto.KumaResource) {
				Expect(krs).To(HaveLen(1))
			})).
			Exec(kds_verifier.Ack(node, mesh.FaultInjectionType)).
			Exec(kds_verifier.ExpectNoResponseDuring(200 * time.Millisecond)).
			Exec(kds_verifier.CloseStream())

		err := vrf.Verify(tc)
		Expect(err).ToNot(HaveOccurred())

		<-tc.Stop()

	})
})
