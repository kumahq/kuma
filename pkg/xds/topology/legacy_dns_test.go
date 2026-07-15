package topology_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	mesh_proto "github.com/kumahq/kuma/v3/api/mesh/v1alpha1"
	system_proto "github.com/kumahq/kuma/v3/api/system/v1alpha1"
	hostnamegenerator_api "github.com/kumahq/kuma/v3/pkg/core/resources/apis/hostnamegenerator/api/v1alpha1"
	meshexternalservice_api "github.com/kumahq/kuma/v3/pkg/core/resources/apis/meshexternalservice/api/v1alpha1"
	meshmzservice_api "github.com/kumahq/kuma/v3/pkg/core/resources/apis/meshmultizoneservice/api/v1alpha1"
	meshservice_api "github.com/kumahq/kuma/v3/pkg/core/resources/apis/meshservice/api/v1alpha1"
	system_api "github.com/kumahq/kuma/v3/pkg/core/resources/apis/system"
	xds_types "github.com/kumahq/kuma/v3/pkg/core/xds/types"
	"github.com/kumahq/kuma/v3/pkg/test/resources/builders"
	xds_topology "github.com/kumahq/kuma/v3/pkg/xds/topology"
)

var _ = Describe("LegacyDomains", func() {
	It("should rebuild legacy aliases from MeshService-family resources", func() {
		kubeMeshService := builders.MeshService().
			WithName("backend").
			WithNamespace("test-ns").
			WithLabels(map[string]string{
				mesh_proto.EnvTag:           mesh_proto.KubernetesEnvironment,
				mesh_proto.KubeNamespaceTag: "test-ns",
			}).
			AddIntPort(8080, 8080, "").
			WithKumaVIP("240.0.0.10").
			Build()
		kubeMeshService.Status.Addresses = []hostnamegenerator_api.Address{{
			Hostname: "backend.test-ns.svc.zone-1.mesh.local",
		}}

		universalMeshService := builders.MeshService().
			WithName("httpbin").
			WithLabels(map[string]string{
				mesh_proto.EnvTag: mesh_proto.UniversalEnvironment,
			}).
			AddIntPort(80, 80, "").
			WithKumaVIP("240.0.0.11").
			Build()

		meshExternalService := builders.MeshExternalService().
			WithName("external-service").
			WithKumaVIP("240.0.0.12").
			Build()

		meshMultiZoneService := builders.MeshMultiZoneService().
			WithName("global-svc").
			WithServiceLabelSelector(map[string]string{"app": "global-svc"}).
			AddPort(meshmzservice_api.Port{Port: 80, AppProtocol: "tcp"}).
			Build()
		meshMultiZoneService.Status.VIPs = []meshservice_api.VIP{{IP: "240.0.0.13"}}

		Expect(xds_topology.LegacyDomains(
			[]*meshservice_api.MeshServiceResource{kubeMeshService, universalMeshService},
			[]*meshexternalservice_api.MeshExternalServiceResource{meshExternalService},
			[]*meshmzservice_api.MeshMultiZoneServiceResource{meshMultiZoneService},
		)).To(Equal([]xds_types.VIPDomains{
			{
				Address: "240.0.0.10",
				Domains: []string{
					"backend.test-ns.svc.8080.mesh",
					"backend_test-ns_svc_8080.mesh",
				},
			},
			{
				Address: "240.0.0.11",
				Domains: []string{"httpbin.mesh"},
			},
			{
				Address: "240.0.0.12",
				Domains: []string{"external-service.mesh"},
			},
			{
				Address: "240.0.0.13",
				Domains: []string{"global-svc.mesh"},
			},
		}))
	})

	It("should remap MeshService hostnames to legacy service VIPs from config", func() {
		meshService := builders.MeshService().
			WithName("local-test-server").
			WithNamespace("mesh-service-reachable-backends").
			WithLabels(map[string]string{
				mesh_proto.EnvTag:           mesh_proto.KubernetesEnvironment,
				mesh_proto.KubeNamespaceTag: "mesh-service-reachable-backends",
			}).
			AddIntPort(80, 80, "").
			WithKumaVIP("241.0.0.10").
			Build()
		meshService.Status.Addresses = []hostnamegenerator_api.Address{{
			Hostname: "local-test-server.mesh-service-reachable-backends.svc.kuma-1.mesh.local",
		}}

		config := system_api.NewConfigResource()
		config.Spec = &system_proto.Config{
			Config: `{"0:local-test-server_mesh-service-reachable-backends_svc_80":{"address":"240.0.0.10","outbounds":[{"TagSet":{"kuma.io/service":"local-test-server_mesh-service-reachable-backends_svc_80"}}]}}`,
		}

		domains, outbounds, err := xds_topology.LegacyVIPCompatibility(
			[]*system_api.ConfigResource{config},
			"mesh",
			80,
			[]*meshservice_api.MeshServiceResource{meshService},
			nil,
			nil,
		)
		Expect(err).ToNot(HaveOccurred())
		Expect(domains).To(ContainElements(
			xds_types.VIPDomains{
				Address: "240.0.0.10",
				Domains: []string{
					"local-test-server.mesh-service-reachable-backends.svc.kuma-1.mesh.local",
				},
			},
			xds_types.VIPDomains{
				Address: "240.0.0.10",
				Domains: []string{
					"local-test-server_mesh-service-reachable-backends_svc_80.mesh",
					"local-test-server.mesh-service-reachable-backends.svc.80.mesh",
				},
			},
		))
		Expect(outbounds).To(ContainElement(&xds_types.Outbound{
			LegacyOutbound: &mesh_proto.Dataplane_Networking_Outbound{
				Address: "240.0.0.10",
				Port:    80,
				Tags: map[string]string{
					mesh_proto.ServiceTag: "local-test-server_mesh-service-reachable-backends_svc_80",
				},
			},
		}))
	})

	It("should ignore host and FQDN entries from persisted virtual outbounds", func() {
		config := system_api.NewConfigResource()
		config.Spec = &system_proto.Config{
			Config: `{"0:local-test-server_mesh-service-reachable-backends_svc_80":{"address":"240.0.0.10","outbounds":[{"TagSet":{"kuma.io/service":"local-test-server_mesh-service-reachable-backends_svc_80"}}]},"1:api.example.com":{"address":"240.0.0.20","outbounds":[{"Port":443,"TagSet":{"kuma.io/service":"external"}}]},"2:echo.internal":{"address":"240.0.0.21","outbounds":[{"Port":8080,"TagSet":{"kuma.io/service":"echo"}}]}}`,
		}

		domains, outbounds, err := xds_topology.LegacyVIPCompatibility(
			[]*system_api.ConfigResource{config},
			"mesh",
			80,
			nil,
			nil,
			nil,
		)
		Expect(err).ToNot(HaveOccurred())
		Expect(domains).To(ConsistOf(xds_types.VIPDomains{
			Address: "240.0.0.10",
			Domains: []string{
				"local-test-server_mesh-service-reachable-backends_svc_80.mesh",
				"local-test-server.mesh-service-reachable-backends.svc.80.mesh",
			},
		}))
		Expect(outbounds).To(ConsistOf(&xds_types.Outbound{
			LegacyOutbound: &mesh_proto.Dataplane_Networking_Outbound{
				Address: "240.0.0.10",
				Port:    80,
				Tags: map[string]string{
					mesh_proto.ServiceTag: "local-test-server_mesh-service-reachable-backends_svc_80",
				},
			},
		}))
	})
})
