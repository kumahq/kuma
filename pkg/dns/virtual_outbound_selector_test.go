package dns_test

import (
	"strconv"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	rest_v1alpha1 "github.com/kumahq/kuma/pkg/core/resources/model/rest/v1alpha1"
	"github.com/kumahq/kuma/pkg/dns"
)

func vob(idx int, selectors ...map[string]string) *core_mesh.VirtualOutboundResource {
	sels := []*mesh_proto.Selector{}
	for _, v := range selectors {
		sels = append(sels, &mesh_proto.Selector{Match: v})
	}
	return &core_mesh.VirtualOutboundResource{
		Meta: &rest_v1alpha1.ResourceMeta{
			Name:         strconv.Itoa(idx),
			CreationTime: time.Date(2021, 12, 12, idx, 0, 0, 0, time.UTC),
		},
		Spec: &mesh_proto.VirtualOutbound{
			Selectors: sels,
		},
	}
}

type voMatchTestCase struct {
	givenVos    []*core_mesh.VirtualOutboundResource
	whenTags    map[string]string
	thenIndices []int
}

var _ = DescribeTable("virtual-outbound match",
	func(tc voMatchTestCase) {

		// When
		res := dns.Match(tc.givenVos, tc.whenTags)

		// Then
		indices := []int{}
		for _, r := range res {
			i, _ := strconv.Atoi(r.Meta.GetName())
			indices = append(indices, i)
		}

		Expect(indices).To(Equal(tc.thenIndices))
	},
	Entry("empty", voMatchTestCase{thenIndices: []int{}}),
	Entry("noMatch", voMatchTestCase{
		givenVos: []*core_mesh.VirtualOutboundResource{
			vob(0, map[string]string{mesh_proto.ServiceTag: "a-service"}),
		},
		whenTags: map[string]string{
			mesh_proto.ServiceTag: "b-service",
		},
		thenIndices: []int{},
	}),
	Entry("matchBoth respect order", voMatchTestCase{
		givenVos: []*core_mesh.VirtualOutboundResource{
			vob(0, map[string]string{mesh_proto.ServiceTag: "*"}),
			vob(1, map[string]string{mesh_proto.ServiceTag: "*", "foo": "bar"}),
		},
		whenTags: map[string]string{
			mesh_proto.ServiceTag: "b-service",
			"foo":                 "bar",
		},
		thenIndices: []int{1, 0},
	}),
	Entry("selector that doesn't match doesn't influence weight", voMatchTestCase{
		givenVos: []*core_mesh.VirtualOutboundResource{
			vob(0, map[string]string{mesh_proto.ServiceTag: "*"}, map[string]string{mesh_proto.ServiceTag: "*", "foo": "baz", "bim": "bam"}),
			vob(1, map[string]string{mesh_proto.ServiceTag: "*", "foo": "bar"}),
		},
		whenTags: map[string]string{
			mesh_proto.ServiceTag: "b-service",
			"foo":                 "bar",
		},
		thenIndices: []int{1, 0},
	}),
	Entry("same weight picks newest", voMatchTestCase{
		givenVos: []*core_mesh.VirtualOutboundResource{
			vob(0, map[string]string{mesh_proto.ServiceTag: "*"}),
			vob(1, map[string]string{mesh_proto.ServiceTag: "*"}),
		},
		whenTags: map[string]string{
			mesh_proto.ServiceTag: "b-service",
			"foo":                 "bar",
		},
		thenIndices: []int{1, 0},
	}),
)
