package generator_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	mesh_proto "github.com/kumahq/kuma/v3/api/mesh/v1alpha1"
	observability_v1 "github.com/kumahq/kuma/v3/api/observability/v1"
	core_mesh "github.com/kumahq/kuma/v3/pkg/core/resources/apis/mesh"
	core_xds "github.com/kumahq/kuma/v3/pkg/core/xds"
	. "github.com/kumahq/kuma/v3/pkg/mads/v1/generator"
	"github.com/kumahq/kuma/v3/pkg/plugins/policies/meshmetric/api/v1alpha1"
	test_model "github.com/kumahq/kuma/v3/pkg/test/resources/model"
	"github.com/kumahq/kuma/v3/pkg/util/pointer"
)

var _ = Describe("Generate()", func() {
	dataplane := &core_mesh.DataplaneResource{
		Meta: &test_model.ResourceMeta{
			Name: "backend-01",
			Mesh: "demo",
		},
		Spec: &mesh_proto.Dataplane{
			Networking: &mesh_proto.Dataplane_Networking{
				Address: "192.168.0.1",
				Inbound: []*mesh_proto.Dataplane_Networking_Inbound{{
					Port:        80,
					ServicePort: 8080,
					Tags: map[string]string{
						"kuma.io/service": "backend",
					},
				}},
			},
		},
	}

	prometheusConf := func(clientId *string) *v1alpha1.Conf {
		return &v1alpha1.Conf{
			Backends: &[]v1alpha1.Backend{{
				Type: v1alpha1.PrometheusBackendType,
				Prometheus: &v1alpha1.PrometheusBackend{
					ClientId: clientId,
					Port:     5670,
					Path:     "/metrics",
				},
			}},
		}
	}

	It("should generate an assignment for a matching Prometheus backend", func() {
		// given
		meshMetricToDataplane := map[*v1alpha1.Conf]*core_mesh.DataplaneResource{
			prometheusConf(nil): dataplane,
		}

		// when
		resources, err := Generate(meshMetricToDataplane, DefaultKumaClientId, false)

		// then
		Expect(err).ToNot(HaveOccurred())
		Expect(resources).To(Equal([]*core_xds.Resource{{
			Name: "/meshes/demo/dataplanes/backend-01",
			Resource: &observability_v1.MonitoringAssignment{
				Mesh:    "demo",
				Service: "backend",
				Targets: []*observability_v1.MonitoringAssignment_Target{{
					Scheme:      "http",
					Name:        "backend-01",
					Address:     "192.168.0.1:5670",
					MetricsPath: "/metrics",
					Labels:      map[string]string{},
				}},
			},
		}}))
	})

	It("should skip a backend with a non-matching clientId", func() {
		// given
		meshMetricToDataplane := map[*v1alpha1.Conf]*core_mesh.DataplaneResource{
			prometheusConf(pointer.To("other-client")): dataplane,
		}

		// when
		resources, err := Generate(meshMetricToDataplane, DefaultKumaClientId, false)

		// then
		Expect(err).ToNot(HaveOccurred())
		Expect(resources).To(BeEmpty())
	})

	It("should skip a non-Prometheus backend", func() {
		// given
		meshMetricToDataplane := map[*v1alpha1.Conf]*core_mesh.DataplaneResource{
			{
				Backends: &[]v1alpha1.Backend{{
					Type: v1alpha1.OpenTelemetryBackendType,
				}},
			}: dataplane,
		}

		// when
		resources, err := Generate(meshMetricToDataplane, DefaultKumaClientId, false)

		// then
		Expect(err).ToNot(HaveOccurred())
		Expect(resources).To(BeEmpty())
	})
})
