package generator_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	mesh_proto "github.com/kumahq/kuma/v2/api/mesh/v1alpha1"
	observability_v1 "github.com/kumahq/kuma/v2/api/observability/v1"
	core_mesh "github.com/kumahq/kuma/v2/pkg/core/resources/apis/mesh"
	. "github.com/kumahq/kuma/v2/pkg/mads/v1/generator"
	"github.com/kumahq/kuma/v2/pkg/plugins/policies/meshmetric/api/v1alpha1"
	test_model "github.com/kumahq/kuma/v2/pkg/test/resources/model"
)

var _ = Describe("Generate() for MeshMetric", func() {
	dataplane := &core_mesh.DataplaneResource{
		Meta: &test_model.ResourceMeta{Name: "backend-01", Mesh: "demo"},
		Spec: &mesh_proto.Dataplane{
			Networking: &mesh_proto.Dataplane_Networking{
				Address: "192.168.0.1",
				Inbound: []*mesh_proto.Dataplane_Networking_Inbound{{
					Port:        80,
					ServicePort: 8080,
					Tags:        map[string]string{mesh_proto.ServiceTag: "backend"},
				}},
			},
		},
	}

	DescribeTable("should select the scheme from the TLS mode",
		func(mode v1alpha1.TlsMode, scheme string) {
			// given
			conf := &v1alpha1.Conf{
				Backends: &[]v1alpha1.Backend{{
					Type: v1alpha1.PrometheusBackendType,
					Prometheus: &v1alpha1.PrometheusBackend{
						Port: 5670,
						Path: "/metrics",
						Tls:  &v1alpha1.PrometheusTls{Mode: mode},
					},
				}},
			}

			// when
			resources, err := Generate(map[*v1alpha1.Conf]*core_mesh.DataplaneResource{conf: dataplane}, DefaultKumaClientId, false)

			// then
			Expect(err).ToNot(HaveOccurred())
			Expect(resources).To(HaveLen(1))
			Expect(resources[0].Resource.(*observability_v1.MonitoringAssignment).Targets[0].Scheme).To(Equal(scheme))
		},
		Entry("Disabled", v1alpha1.Disabled, "http"),
		Entry("ProvidedTLS", v1alpha1.ProvidedTLS, "https"),
		Entry("ActiveMTLSBackend", v1alpha1.ActiveMTLSBackend, "https"),
	)
})
