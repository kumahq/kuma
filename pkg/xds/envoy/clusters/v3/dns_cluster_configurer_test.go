package clusters_test

import (
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
	"github.com/kumahq/kuma/pkg/xds/envoy"
	"github.com/kumahq/kuma/pkg/xds/envoy/clusters"
)

var _ = Describe("DNSClusterConfigurer", func() {

	It("should generate proper Envoy config", func() {
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
        name: test:cluster
        type: STRICT_DNS`

		// when
		cluster, err := clusters.NewClusterBuilder(envoy.APIV3).
			Configure(clusters.DNSCluster(clusterName, address, port)).
			Configure(clusters.Timeout(core_mesh.ProtocolTCP, &mesh_proto.Timeout_Conf{ConnectTimeout: util_proto.Duration(5 * time.Second)})).
			Build()

		// then
		Expect(err).ToNot(HaveOccurred())

		actual, err := util_proto.ToYAML(cluster)
		Expect(err).ToNot(HaveOccurred())
		Expect(actual).To(MatchYAML(expected))
	})
})
