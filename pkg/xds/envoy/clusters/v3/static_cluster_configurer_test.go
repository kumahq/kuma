package clusters_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
	"github.com/kumahq/kuma/pkg/xds/envoy"
	"github.com/kumahq/kuma/pkg/xds/envoy/clusters"
)

var _ = Describe("StaticClusterConfigurer", func() {

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
			Configure(clusters.StaticCluster(clusterName, address, port)).
			Configure(clusters.Timeout(core_mesh.ProtocolTCP, DefaultTimeout())).
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
			Configure(clusters.StaticClusterUnixSocket(clusterName, path)).
			Configure(clusters.Timeout(core_mesh.ProtocolTCP, DefaultTimeout())).
			Build()

		// then
		Expect(err).ToNot(HaveOccurred())

		actual, err := util_proto.ToYAML(cluster)
		Expect(err).ToNot(HaveOccurred())
		Expect(actual).To(MatchYAML(expected))
	})
})
