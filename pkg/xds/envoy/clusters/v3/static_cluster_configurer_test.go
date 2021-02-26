package clusters_test

import (
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"google.golang.org/protobuf/types/known/durationpb"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	mesh_core "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"

	util_proto "github.com/kumahq/kuma/pkg/util/proto"
	"github.com/kumahq/kuma/pkg/xds/envoy"
	"github.com/kumahq/kuma/pkg/xds/envoy/clusters"
)

var _ = Describe("StaticClusterConfigurer", func() {

	It("should generate proper Envoy config", func() {
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
			Configure(clusters.Timeout(mesh_core.ProtocolTCP, &mesh_proto.Timeout_Conf{ConnectTimeout: durationpb.New(5 * time.Second)})).
			Build()

		// then
		Expect(err).ToNot(HaveOccurred())

		actual, err := util_proto.ToYAML(cluster)
		Expect(err).ToNot(HaveOccurred())
		Expect(actual).To(MatchYAML(expected))
	})
})
