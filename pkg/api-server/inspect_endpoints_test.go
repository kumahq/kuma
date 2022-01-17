package api_server_test

import (
	"context"
	"io"
	"net/http"
	"net/url"
	"path"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	"google.golang.org/protobuf/types/known/durationpb"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	config "github.com/kumahq/kuma/pkg/config/api-server"
	"github.com/kumahq/kuma/pkg/core"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/manager"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/store"
	"github.com/kumahq/kuma/pkg/metrics"
	"github.com/kumahq/kuma/pkg/plugins/resources/memory"
	"github.com/kumahq/kuma/pkg/test/kds/samples"
	"github.com/kumahq/kuma/pkg/test/matchers"
	test_model "github.com/kumahq/kuma/pkg/test/resources/model"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
)

type dataplaneBuilder core_mesh.DataplaneResource

func newMesh(name string) *core_mesh.MeshResource {
	return &core_mesh.MeshResource{
		Meta: &test_model.ResourceMeta{Name: name},
		Spec: &mesh_proto.Mesh{},
	}
}

func newDataplaneBuilder() *dataplaneBuilder {
	return &dataplaneBuilder{
		Spec: &mesh_proto.Dataplane{
			Networking: &mesh_proto.Dataplane_Networking{
				Address: "1.1.1.1",
			},
		},
	}
}

func (b *dataplaneBuilder) build() *core_mesh.DataplaneResource {
	return (*core_mesh.DataplaneResource)(b)
}

func (b *dataplaneBuilder) meta(name, mesh string) *dataplaneBuilder {
	b.Meta = &test_model.ResourceMeta{Name: name, Mesh: mesh}
	return b
}

func (b *dataplaneBuilder) inbound(service, ip string, dpPort, workloadPort uint32) *dataplaneBuilder {
	b.Spec.Networking.Inbound = append(b.Spec.Networking.Inbound, &mesh_proto.Dataplane_Networking_Inbound{
		Address:     ip,
		Port:        dpPort,
		ServicePort: workloadPort,
		Tags: map[string]string{
			mesh_proto.ServiceTag:  service,
			mesh_proto.ProtocolTag: "http",
		},
	})
	return b
}

func (b *dataplaneBuilder) outbound(service, ip string, port uint32) *dataplaneBuilder {
	b.Spec.Networking.Outbound = append(b.Spec.Networking.Outbound, &mesh_proto.Dataplane_Networking_Outbound{
		Address: ip,
		Port:    port,
		Tags: map[string]string{
			mesh_proto.ServiceTag: service,
		},
	})
	return b
}

func serviceSelector(service, protocol string) []*mesh_proto.Selector {
	selector := &mesh_proto.Selector{
		Match: map[string]string{
			mesh_proto.ServiceTag: service,
		},
	}
	if protocol != "" {
		selector.Match[mesh_proto.ProtocolTag] = protocol
	}
	return []*mesh_proto.Selector{selector}
}

var _ = Describe("Inspect WS", func() {

	type testCase struct {
		path       string
		goldenFile string
		resources  []core_model.Resource
	}

	DescribeTable("should return policies matched for specific dataplane",
		func(given testCase) {
			// setup
			resourceStore := memory.NewStore()
			metrics, err := metrics.NewMetrics("Standalone")
			Expect(err).ToNot(HaveOccurred())

			core.Now = func() time.Time { return time.Time{} }

			rm := manager.NewResourceManager(resourceStore)
			for _, resource := range given.resources {
				err = rm.Create(context.Background(), resource,
					store.CreateBy(core_model.MetaToResourceKey(resource.GetMeta())))
				Expect(err).ToNot(HaveOccurred())
			}

			apiServer := createTestApiServer(resourceStore, config.DefaultApiServerConfig(), true, metrics)

			stop := make(chan struct{})
			go func() {
				defer GinkgoRecover()
				err := apiServer.Start(stop)
				Expect(err).ToNot(HaveOccurred())
			}()

			// when
			var resp *http.Response
			Eventually(func() error {
				r, err := http.Get((&url.URL{
					Scheme: "http",
					Host:   apiServer.Address(),
					Path:   given.path,
				}).String())
				resp = r
				return err
			}, "3s").ShouldNot(HaveOccurred())

			// then
			bytes, err := io.ReadAll(resp.Body)
			Expect(err).ToNot(HaveOccurred())
			Expect(bytes).To(matchers.MatchGoldenJSON(path.Join("testdata", given.goldenFile)))
		},
		Entry("full example", testCase{
			path:       "meshes/default/dataplanes/backend-1/policies",
			goldenFile: "inspect.json",
			resources: []core_model.Resource{
				newMesh("default"),
				newDataplaneBuilder().
					meta("backend-1", "default").
					inbound("backend", "192.168.0.1", 80, 81).
					outbound("redis", "192.168.0.2", 8080).
					outbound("gateway", "192.168.0.3", 8080).
					outbound("postgres", "192.168.0.4", 8080).
					outbound("web", "192.168.0.2", 8080).
					build(),
				&core_mesh.TrafficPermissionResource{
					Meta: &test_model.ResourceMeta{Name: "tp-1", Mesh: "default"},
					Spec: &mesh_proto.TrafficPermission{
						Sources:      serviceSelector("*", ""),
						Destinations: serviceSelector("*", ""),
					},
				},
				&core_mesh.FaultInjectionResource{
					Meta: &test_model.ResourceMeta{Name: "fi-1", Mesh: "default"},
					Spec: &mesh_proto.FaultInjection{
						Sources:      serviceSelector("*", ""),
						Destinations: serviceSelector("backend", "http"),
						Conf: &mesh_proto.FaultInjection_Conf{
							Delay: &mesh_proto.FaultInjection_Conf_Delay{
								Value:      durationpb.New(5 * time.Second),
								Percentage: util_proto.Double(90),
							},
						},
					},
				},
				&core_mesh.FaultInjectionResource{
					Meta: &test_model.ResourceMeta{Name: "fi-2", Mesh: "default"},
					Spec: &mesh_proto.FaultInjection{
						Sources:      serviceSelector("*", ""),
						Destinations: serviceSelector("backend", "http"),
						Conf: &mesh_proto.FaultInjection_Conf{
							Abort: &mesh_proto.FaultInjection_Conf_Abort{
								HttpStatus: util_proto.UInt32(500),
								Percentage: util_proto.Double(80),
							},
						},
					},
				},
				&core_mesh.TimeoutResource{
					Meta: &test_model.ResourceMeta{Name: "t-1", Mesh: "default"},
					Spec: &mesh_proto.Timeout{
						Sources:      serviceSelector("*", ""),
						Destinations: serviceSelector("redis", ""),
						Conf:         samples.Timeout.Conf,
					},
				},
				&core_mesh.HealthCheckResource{
					Meta: &test_model.ResourceMeta{Name: "hc-1", Mesh: "default"},
					Spec: &mesh_proto.HealthCheck{
						Sources:      serviceSelector("backend", ""),
						Destinations: serviceSelector("*", ""),
						Conf:         samples.HealthCheck.Conf,
					},
				},
			},
		}),
	)
})
