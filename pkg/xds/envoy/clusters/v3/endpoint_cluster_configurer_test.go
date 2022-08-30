package clusters_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
	"github.com/kumahq/kuma/pkg/xds/envoy"
	"github.com/kumahq/kuma/pkg/xds/envoy/clusters"
)

var _ = Describe("ProvidedEndpointClusterConfigurer", func() {

	It("should generate proper Envoy config", func() {
		// given
		clusterName := "test:cluster"
		address := "google.com"
		port := uint32(80)
		expected := `
        altStatName: test_cluster
        connectTimeout: 5s
        dnsLookupFamily: V4_ONLY
        loadAssignment:
          clusterName: test:cluster
          endpoints:
          - lbEndpoints:
            - endpoint:
                address:
                  socketAddress:
                    address: google.com
                    portValue: 80
              loadBalancingWeight: 100
        name: test:cluster
        type: STRICT_DNS`

		// when
		cluster, err := clusters.NewClusterBuilder(envoy.APIV3).
			Configure(clusters.ProvidedEndpointCluster(clusterName, false,
				core_xds.Endpoint{
					Target: address,
					Port:   port,
					Tags:   nil,
					Weight: 100,
					ExternalService: &core_xds.ExternalService{
						TLSEnabled: true,
					},
				},
			)).
			Configure(clusters.Timeout(DefaultTimeout(), core_mesh.ProtocolTCP)).
			Build()

		// then
		Expect(err).ToNot(HaveOccurred())

		actual, err := util_proto.ToYAML(cluster)
		Expect(err).ToNot(HaveOccurred())
		Expect(actual).To(MatchYAML(expected))
	})

	It("should generate proper Envoy config with IPv6", func() {
		// given
		clusterName := "test:cluster"
		address := "google.com"
		port := uint32(80)
		expected := `
        altStatName: test_cluster
        connectTimeout: 5s
        loadAssignment:
          clusterName: test:cluster
          endpoints:
          - lbEndpoints:
            - endpoint:
                address:
                  socketAddress:
                    address: google.com
                    portValue: 80
              loadBalancingWeight: 100
        name: test:cluster
        type: STRICT_DNS`

		// when
		cluster, err := clusters.NewClusterBuilder(envoy.APIV3).
			Configure(clusters.ProvidedEndpointCluster(clusterName, true,
				core_xds.Endpoint{
					Target: address,
					Port:   port,
					Tags:   nil,
					Weight: 100,
					ExternalService: &core_xds.ExternalService{
						TLSEnabled: true,
					},
				},
			)).
			Configure(clusters.Timeout(DefaultTimeout(), core_mesh.ProtocolTCP)).
			Build()

		// then
		Expect(err).ToNot(HaveOccurred())

		actual, err := util_proto.ToYAML(cluster)
		Expect(err).ToNot(HaveOccurred())
		Expect(actual).To(MatchYAML(expected))
	})

	It("should generate proper Envoy config for static cluster with socket", func() {
		// given
		clusterName := "test:cluster"
		address := "192.168.0.1"
		port := uint32(8080)
		expected := `
        altStatName: test_cluster
        connectTimeout: 5s
        loadAssignment:
          clusterName: test:cluster
          endpoints:
          - lbEndpoints:
            - endpoint:
                address:
                  socketAddress:
                    address: 192.168.0.1
                    portValue: 8080
        name: test:cluster
        type: STATIC`

		// when
		cluster, err := clusters.NewClusterBuilder(envoy.APIV3).
			Configure(clusters.ProvidedEndpointCluster(clusterName, false, core_xds.Endpoint{Target: address, Port: port})).
			Configure(clusters.Timeout(DefaultTimeout(), core_mesh.ProtocolTCP)).
			Build()

		// then
		Expect(err).ToNot(HaveOccurred())

		actual, err := util_proto.ToYAML(cluster)
		Expect(err).ToNot(HaveOccurred())
		Expect(actual).To(MatchYAML(expected))
	})

	It("should generate proper Envoy config for static cluster with unix socket", func() {
		// given
		clusterName := "test:cluster"
		path := "/tmp/socket_file_name.sock"
		expected := `
        altStatName: test_cluster
        connectTimeout: 5s
        loadAssignment:
          clusterName: test:cluster
          endpoints:
          - lbEndpoints:
            - endpoint:
                address:
                  pipe:
                    path: /tmp/socket_file_name.sock
        name: test:cluster
        type: STATIC`

		// when
		cluster, err := clusters.NewClusterBuilder(envoy.APIV3).
			Configure(clusters.ProvidedEndpointCluster(clusterName, false, core_xds.Endpoint{UnixDomainPath: path})).
			Configure(clusters.Timeout(DefaultTimeout(), core_mesh.ProtocolTCP)).
			Build()

		// then
		Expect(err).ToNot(HaveOccurred())

		actual, err := util_proto.ToYAML(cluster)
		Expect(err).ToNot(HaveOccurred())
		Expect(actual).To(MatchYAML(expected))
	})
})
