package clusters

import (
	. "github.com/onsi/ginkgo"
)

var _ = Describe("TimeoutConfigurer", func() {

	//It("should generate proper Envoy config", func() {
	//	// given
	//	clusterName := "test:cluster"
	//	address := "192.168.0.1"
	//	port := uint32(8080)
	//	expected := `
    //    altStatName: test_cluster
    //    connectTimeout: 5s
    //    loadAssignment:
    //      clusterName: test:cluster
    //      endpoints:
    //      - lbEndpoints:
    //        - endpoint:
    //            address:
    //              socketAddress:
    //                address: 192.168.0.1
    //                portValue: 8080
    //    name: test:cluster
    //    type: STATIC`
	//
	//	// when
	//	cluster, err := clusters.NewClusterBuilder(envoy.APIV2).
	//		Configure(clusters.StaticCluster(clusterName, address, port)).
	//		Configure(clusters.Timeout(mesh_core.ProtocolTCP, &mesh_proto.Timeout_Conf{ConnectTimeout: durationpb.New(5 * time.Second)})).
	//		Build()
	//
	//	// then
	//	Expect(err).ToNot(HaveOccurred())
	//
	//	actual, err := util_proto.ToYAML(cluster)
	//	Expect(err).ToNot(HaveOccurred())
	//	Expect(actual).To(MatchYAML(expected))
	//})
})
