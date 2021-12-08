package xds_test

import (
	"reflect"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

var _ = Describe("ValidateMatchedPoliciesType", func() {

	type testCase struct {
		typ    reflect.Type
		errMsg string
	}

	DescribeTable("should pass validation",
		func(given testCase) {
			Expect(core_xds.ValidateMatchedPoliciesType(given.typ)).To(Succeed())
		},
		Entry("empty type", testCase{
			typ: reflect.TypeOf(struct {
			}{}),
		}),
		Entry("full example", testCase{
			typ: reflect.TypeOf(struct {
				strToTrafficRoutes      map[string][]*core_mesh.TrafficRouteResource
				inboundToTrafficRoutes  map[mesh_proto.InboundInterface][]*core_mesh.TrafficRouteResource
				outboundToTrafficRoutes map[mesh_proto.OutboundInterface][]*core_mesh.TrafficRouteResource

				strToTrafficRoute      map[string]*core_mesh.TrafficRouteResource
				inboundToTrafficRoute  map[mesh_proto.InboundInterface]*core_mesh.TrafficRouteResource
				outboundToTrafficRoute map[mesh_proto.OutboundInterface]*core_mesh.TrafficRouteResource

				trafficRoute  *core_mesh.TrafficRouteResource
				trafficRoutes []*core_mesh.TrafficRouteResource
			}{}),
		}),
	)

	DescribeTable("should not pass validation",
		func(given testCase) {
			err := core_xds.ValidateMatchedPoliciesType(given.typ)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal(given.errMsg))
		},
		Entry("wrong key type", testCase{
			typ: reflect.TypeOf(struct {
				intToTrafficRoutes map[int][]*core_mesh.TrafficRouteResource
			}{}),
			errMsg: "field intToTrafficRoutes: key has wrong type int",
		}),
		Entry("wrong value type", testCase{
			typ: reflect.TypeOf(struct {
				inboundToTrafficRoute map[mesh_proto.InboundInterface]*mesh_proto.TrafficRoute
			}{}),
			errMsg: "field inboundToTrafficRoute: value doesn't implement Resource",
		}),
		Entry("wrong field type", testCase{
			typ: reflect.TypeOf(struct {
				trafficRoutes []*mesh_proto.TrafficRoute
			}{}),
			errMsg: "field trafficRoutes: value doesn't implement Resource",
		}),
	)
})
