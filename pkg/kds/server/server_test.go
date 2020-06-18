package server_test

import (
	"context"
	"fmt"
	"time"

	v2 "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	envoy_core "github.com/envoyproxy/go-control-plane/envoy/api/v2/core"
	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/duration"
	"github.com/golang/protobuf/ptypes/wrappers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"
	"google.golang.org/grpc"

	mesh_proto "github.com/Kong/kuma/api/mesh/v1alpha1"
	kuma_cp "github.com/Kong/kuma/pkg/config/app/kuma-cp"
	kds_config "github.com/Kong/kuma/pkg/config/kds"
	"github.com/Kong/kuma/pkg/core"
	"github.com/Kong/kuma/pkg/core/resources/apis/mesh"
	"github.com/Kong/kuma/pkg/core/resources/manager"
	"github.com/Kong/kuma/pkg/core/resources/model"
	"github.com/Kong/kuma/pkg/core/resources/store"
	"github.com/Kong/kuma/pkg/core/runtime"
	"github.com/Kong/kuma/pkg/core/runtime/component"
	"github.com/Kong/kuma/pkg/kds"
	kds_server "github.com/Kong/kuma/pkg/kds/server"
	mads_server "github.com/Kong/kuma/pkg/mads/server"
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

type mockStream struct {
	ctx       context.Context
	recv      chan *v2.DiscoveryRequest
	sent      chan *v2.DiscoveryResponse
	nonce     int
	sendError bool
	grpc.ServerStream
}

func (stream *mockStream) Context() context.Context {
	return stream.ctx
}

func (stream *mockStream) Send(resp *v2.DiscoveryResponse) error {
	// check that nonce is monotonically incrementing
	stream.nonce++
	Expect(resp.Nonce).To(Equal(fmt.Sprintf("%d", stream.nonce)))
	Expect(resp.VersionInfo).ToNot(BeEmpty())
	Expect(resp.Resources).ToNot(BeEmpty())
	Expect(resp.TypeUrl).ToNot(BeEmpty())
	for _, res := range resp.Resources {
		Expect(res.TypeUrl).To(Equal(resp.TypeUrl))
	}

	stream.sent <- resp
	if stream.sendError {
		return errors.New("send error")
	}
	return nil
}

func (stream *mockStream) Recv() (*v2.DiscoveryRequest, error) {
	req, more := <-stream.recv
	if !more {
		return nil, errors.New("empty")
	}
	return req, nil
}

func makeMockStream() *mockStream {
	return &mockStream{
		ctx:  context.Background(),
		sent: make(chan *v2.DiscoveryResponse, 10),
		recv: make(chan *v2.DiscoveryRequest, 10),
	}
}

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

type fn func(ctx testContext) error

func create(ctx context.Context, r model.Resource, opts ...store.CreateOptionsFunc) fn {
	return func(tc testContext) error {
		return tc.store().Create(ctx, r, opts...)
	}
}

func discoveryRequest(resourceType model.ResourceType) fn {
	return func(tc testContext) error {
		tc.stream().recv <- &v2.DiscoveryRequest{
			Node:    node,
			TypeUrl: kds.TypeURL(resourceType),
		}
		return nil
	}
}

func ack(resourceType model.ResourceType) fn {
	return func(tc testContext) error {
		tc.stream().recv <- &v2.DiscoveryRequest{
			Node:          node,
			TypeUrl:       kds.TypeURL(resourceType),
			ResponseNonce: tc.lastResponse(kds.TypeURL(resourceType)).Nonce,
			VersionInfo:   tc.lastResponse(kds.TypeURL(resourceType)).VersionInfo,
		}
		return nil
	}
}

func waitResponse(timeout time.Duration, testFunc func(krs []*mesh_proto.KumaResource)) fn {
	return func(tc testContext) error {
		select {
		case resp := <-tc.stream().sent:
			krs, err := kumaResources(resp)
			if err != nil {
				return err
			}
			if len(krs) > 0 {
				tc.saveLastResponse(krs[0].Spec.TypeUrl, resp)
			}
			testFunc(krs)
		case <-time.After(timeout):
			return fmt.Errorf("timeout exceeded")
		}
		return nil
	}
}

func closeStream() fn {
	return func(tc testContext) error {
		close(tc.stream().recv)
		return nil
	}
}

type verifier interface {
	exec(fn) verifier
	verify(testContext) error
}

func newVerifier() verifier {
	return &verifierImpl{}
}

type verifierImpl struct {
	fns []fn
}

func (v *verifierImpl) exec(f fn) verifier {
	v.fns = append(v.fns, f)
	return v
}

func (v *verifierImpl) verify(tc testContext) error {
	for _, f := range v.fns {
		if err := f(tc); err != nil {
			return err
		}
	}
	return nil
}

type testContext interface {
	store() store.ResourceStore
	stream() *mockStream
	stop() chan struct{}
	saveLastResponse(typ string, response *v2.DiscoveryResponse)
	lastResponse(typeURL string) *v2.DiscoveryResponse
}

type testContextImpl struct {
	resourceStore store.ResourceStore
	mockStream    *mockStream
	stopCh        chan struct{}
	responses     map[string]*v2.DiscoveryResponse
}

func (t *testContextImpl) store() store.ResourceStore {
	return t.resourceStore
}

func (t *testContextImpl) stream() *mockStream {
	return t.mockStream
}

func (t *testContextImpl) stop() chan struct{} {
	return t.stopCh
}

func (t *testContextImpl) saveLastResponse(typ string, response *v2.DiscoveryResponse) {
	t.responses[typ] = response
}

func (t *testContextImpl) lastResponse(typ string) *v2.DiscoveryResponse {
	return t.responses[typ]
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
)

func kumaResources(response *v2.DiscoveryResponse) (resources []*mesh_proto.KumaResource, _ error) {
	for _, r := range response.Resources {
		kr := &mesh_proto.KumaResource{}
		if err := ptypes.UnmarshalAny(r, kr); err != nil {
			return nil, err
		}
		resources = append(resources, kr)
	}
	return
}

var _ = Describe("KDS Server", func() {
	createServer := func(rt runtime.Runtime) kds_server.Server {
		log := core.Log
		hasher, cache := mads_server.NewXdsContext(log)
		generator := kds_server.NewSnapshotGenerator(rt)
		versioner := mads_server.NewVersioner()
		reconciler := mads_server.NewReconciler(hasher, cache, generator, versioner)
		syncTracker := mads_server.NewSyncTracker(reconciler, rt.Config().KdsServer.RefreshInterval)
		callbacks := util_xds.CallbacksChain{
			util_xds.LoggingCallbacks{Log: log},
			syncTracker,
		}
		return kds_server.NewServer(cache, callbacks, log)
	}

	var tc testContext
	BeforeEach(func() {
		s := memory.NewStore()
		mgr := manager.NewResourceManager(s)
		srv := createServer(&testRuntimeContext{
			rom: mgr,
			cfg: kuma_cp.Config{
				KdsServer: &kds_config.KumaDiscoveryServerConfig{
					RefreshInterval: 1 * time.Second,
				},
			},
		})
		stream := makeMockStream()
		stop := make(chan struct{})
		go func() {
			err := srv.StreamKumaResources(stream)
			Expect(err).ToNot(HaveOccurred())
			close(stop)
		}()

		tc = &testContextImpl{
			resourceStore: s,
			mockStream:    stream,
			stopCh:        stop,
			responses:     map[string]*v2.DiscoveryResponse{},
		}
	})

	It("should support all existing resource types", func() {
		ctx := context.Background()

		vrf := newVerifier().
			exec(create(ctx, &mesh.MeshResource{Spec: mesh1}, store.CreateByKey("mesh-1", "mesh-1"))).
			exec(create(ctx, &mesh.DataplaneResource{Spec: ingress}, store.CreateByKey("ingress-1", "mesh-1"))).
			exec(create(ctx, &mesh.CircuitBreakerResource{Spec: cb1}, store.CreateByKey("cb-1", "mesh-1"))).
			exec(create(ctx, &mesh.FaultInjectionResource{Spec: fi1}, store.CreateByKey("fi-1", "mesh-1"))).
			exec(create(ctx, &mesh.HealthCheckResource{Spec: hc1}, store.CreateByKey("hc-1", "mesh-1"))).
			exec(create(ctx, &mesh.TrafficLogResource{Spec: tl1}, store.CreateByKey("tl-1", "mesh-1"))).
			exec(create(ctx, &mesh.TrafficPermissionResource{Spec: tp1}, store.CreateByKey("tp-1", "mesh-1"))).
			exec(create(ctx, &mesh.TrafficRouteResource{Spec: tr1}, store.CreateByKey("tr-1", "mesh-1"))).
			exec(create(ctx, &mesh.TrafficTraceResource{Spec: tt1}, store.CreateByKey("tt-1", "mesh-1"))).
			exec(create(ctx, &mesh.ProxyTemplateResource{Spec: pt1}, store.CreateByKey("pt-1", "mesh-1"))).
			exec(discoveryRequest(mesh.MeshType)).
			exec(waitResponse(defaultTimeout, func(krs []*mesh_proto.KumaResource) {
				Expect(krs).To(HaveLen(1))
				m := &mesh_proto.Mesh{}
				Expect(ptypes.UnmarshalAny(krs[0].Spec, m)).ToNot(HaveOccurred())
				Expect(m).To(Equal(&mesh1))
			})).
			exec(discoveryRequest(mesh.DataplaneType)).
			exec(waitResponse(defaultTimeout, func(krs []*mesh_proto.KumaResource) {
				Expect(krs).To(HaveLen(1))
				m := &mesh_proto.Dataplane{}
				Expect(ptypes.UnmarshalAny(krs[0].Spec, m)).ToNot(HaveOccurred())
				Expect(m).To(Equal(&ingress))
			})).
			exec(discoveryRequest(mesh.CircuitBreakerType)).
			exec(waitResponse(defaultTimeout, func(krs []*mesh_proto.KumaResource) {
				Expect(krs).To(HaveLen(1))
				m := &mesh_proto.CircuitBreaker{}
				Expect(ptypes.UnmarshalAny(krs[0].Spec, m)).ToNot(HaveOccurred())
				Expect(m).To(Equal(&cb1))
			})).
			exec(discoveryRequest(mesh.FaultInjectionType)).
			exec(waitResponse(defaultTimeout, func(krs []*mesh_proto.KumaResource) {
				Expect(krs).To(HaveLen(1))
				m := &mesh_proto.FaultInjection{}
				Expect(ptypes.UnmarshalAny(krs[0].Spec, m)).ToNot(HaveOccurred())
				Expect(m).To(Equal(&fi1))
			})).
			exec(discoveryRequest(mesh.HealthCheckType)).
			exec(waitResponse(defaultTimeout, func(krs []*mesh_proto.KumaResource) {
				Expect(krs).To(HaveLen(1))
				m := &mesh_proto.HealthCheck{}
				Expect(ptypes.UnmarshalAny(krs[0].Spec, m)).ToNot(HaveOccurred())
				Expect(m).To(Equal(&hc1))
			})).
			exec(discoveryRequest(mesh.TrafficLogType)).
			exec(waitResponse(defaultTimeout, func(krs []*mesh_proto.KumaResource) {
				Expect(krs).To(HaveLen(1))
				m := &mesh_proto.TrafficLog{}
				Expect(ptypes.UnmarshalAny(krs[0].Spec, m)).ToNot(HaveOccurred())
				Expect(m).To(Equal(&tl1))
			})).
			exec(discoveryRequest(mesh.TrafficPermissionType)).
			exec(waitResponse(defaultTimeout, func(krs []*mesh_proto.KumaResource) {
				Expect(krs).To(HaveLen(1))
				m := &mesh_proto.TrafficPermission{}
				Expect(ptypes.UnmarshalAny(krs[0].Spec, m)).ToNot(HaveOccurred())
				Expect(m).To(Equal(&tp1))
			})).
			exec(discoveryRequest(mesh.TrafficRouteType)).
			exec(waitResponse(defaultTimeout, func(krs []*mesh_proto.KumaResource) {
				Expect(krs).To(HaveLen(1))
				m := &mesh_proto.TrafficRoute{}
				Expect(ptypes.UnmarshalAny(krs[0].Spec, m)).ToNot(HaveOccurred())
				Expect(m).To(Equal(&tr1))
			})).
			exec(discoveryRequest(mesh.TrafficTraceType)).
			exec(waitResponse(defaultTimeout, func(krs []*mesh_proto.KumaResource) {
				Expect(krs).To(HaveLen(1))
				m := &mesh_proto.TrafficTrace{}
				Expect(ptypes.UnmarshalAny(krs[0].Spec, m)).ToNot(HaveOccurred())
				Expect(m).To(Equal(&tt1))
			})).
			exec(discoveryRequest(mesh.ProxyTemplateType)).
			exec(waitResponse(defaultTimeout, func(krs []*mesh_proto.KumaResource) {
				Expect(krs).To(HaveLen(1))
				m := &mesh_proto.ProxyTemplate{}
				Expect(ptypes.UnmarshalAny(krs[0].Spec, m)).ToNot(HaveOccurred())
				Expect(m).To(Equal(&pt1))
			})).
			exec(closeStream())

		err := vrf.verify(tc)
		Expect(err).ToNot(HaveOccurred())

		<-tc.stop()
	})

	It("should accept request independently for each type", func() {
		ctx := context.Background()

		vrf := newVerifier().
			exec(create(ctx, &mesh.MeshResource{Spec: mesh1}, store.CreateByKey("mesh-1", "mesh-1"))).
			exec(create(ctx, &mesh.FaultInjectionResource{Spec: fi1}, store.CreateByKey("fi1", "mesh-2"))).
			exec(discoveryRequest(mesh.MeshType)).
			exec(discoveryRequest(mesh.FaultInjectionType)).
			exec(waitResponse(defaultTimeout, func(krs []*mesh_proto.KumaResource) {
				Expect(krs).To(HaveLen(1))
			})).
			exec(ack(mesh.MeshType)).
			exec(waitResponse(defaultTimeout, func(krs []*mesh_proto.KumaResource) {
				Expect(krs).To(HaveLen(1))
			})).
			exec(ack(mesh.FaultInjectionType)).
			exec(closeStream())

		err := vrf.verify(tc)
		Expect(err).ToNot(HaveOccurred())

		<-tc.stop()
	})

	It("should send response for resources created after DiscoveryRequest", func() {
		ctx := context.Background()

		vrf := newVerifier().
			exec(discoveryRequest(mesh.MeshType)).
			exec(create(ctx, &mesh.MeshResource{Spec: mesh1}, store.CreateByKey("mesh-1", "mesh-1"))).
			exec(waitResponse(defaultTimeout, func(krs []*mesh_proto.KumaResource) {
				Expect(krs).To(HaveLen(1))
			})).
			exec(ack(mesh.MeshType)).
			exec(closeStream())

		err := vrf.verify(tc)
		Expect(err).ToNot(HaveOccurred())

		<-tc.stop()
	})

	It("should send response for resources created before ACK", func() {
		ctx := context.Background()

		vrf := newVerifier().
			exec(create(ctx, &mesh.MeshResource{Spec: mesh1}, store.CreateByKey("mesh-1", "mesh-1"))).
			exec(discoveryRequest(mesh.MeshType)).
			exec(waitResponse(defaultTimeout, func(krs []*mesh_proto.KumaResource) {
				Expect(krs).To(HaveLen(1))
			})).
			exec(create(ctx, &mesh.MeshResource{Spec: mesh2}, store.CreateByKey("mesh-2", "mesh-2"))).
			exec(ack(mesh.MeshType)).
			exec(waitResponse(10*time.Second, func(krs []*mesh_proto.KumaResource) {
				Expect(krs).To(HaveLen(2))
			})).
			exec(closeStream())

		err := vrf.verify(tc)
		Expect(err).ToNot(HaveOccurred())

		<-tc.stop()
	})

	It("should support update", func() {
		ctx := context.Background()

		vrf := newVerifier().
			exec(create(ctx, &mesh.MeshResource{Spec: mesh1}, store.CreateByKey("mesh-1", "mesh-1"))).
			exec(discoveryRequest(mesh.MeshType)).
			exec(waitResponse(defaultTimeout, func(krs []*mesh_proto.KumaResource) {
				Expect(krs).To(HaveLen(1))
				m := &mesh_proto.Mesh{}
				Expect(ptypes.UnmarshalAny(krs[0].Spec, m)).ToNot(HaveOccurred())
				Expect(m).To(Equal(&mesh1))
			})).
			exec(ack(mesh.MeshType)).
			exec(func(tc testContext) error {
				var meshRes mesh.MeshResource
				if err := tc.store().Get(ctx, &meshRes, store.GetByKey("mesh-1", "mesh-1")); err != nil {
					return err
				}
				meshRes.Spec = mesh2
				if err := tc.store().Update(ctx, &meshRes); err != nil {
					return err
				}
				return nil
			}).
			exec(waitResponse(defaultTimeout, func(krs []*mesh_proto.KumaResource) {
				Expect(krs).To(HaveLen(1))
				m := &mesh_proto.Mesh{}
				Expect(ptypes.UnmarshalAny(krs[0].Spec, m)).ToNot(HaveOccurred())
				Expect(m).To(Equal(&mesh2))
			})).
			exec(closeStream())

		err := vrf.verify(tc)
		Expect(err).ToNot(HaveOccurred())

		<-tc.stop()

	})
})
