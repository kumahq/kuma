package system_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	system_proto "github.com/kumahq/kuma/v2/api/system/v1alpha1"
	"github.com/kumahq/kuma/v2/pkg/core/resources/apis/system"
	core_model "github.com/kumahq/kuma/v2/pkg/core/resources/model"
	"github.com/kumahq/kuma/v2/pkg/core/resources/model/rest"
	"github.com/kumahq/kuma/v2/pkg/test/resources/model"
	util_proto "github.com/kumahq/kuma/v2/pkg/util/proto"
)

var _ = Describe("ZoneOverview", func() {
	overviewJSON := func(overview core_model.Resource) string {
		GinkgoHelper()
		bytes, err := rest.From.Resource(overview).(interface{ MarshalJSON() ([]byte, error) }).MarshalJSON()
		Expect(err).ToNot(HaveOccurred())
		return string(bytes)
	}

	DescribeTable("should always render a zone, whatever the Zone spec is",
		func(spec *system_proto.Zone, expected string) {
			zone := &system.ZoneResource{
				Meta: &model.ResourceMeta{Name: "zone-1"},
				Spec: spec,
			}

			overview := system.NewZoneOverviewResource()
			Expect(overview.SetOverviewSpec(zone, nil)).To(Succeed())
			Expect(overviewJSON(overview)).To(ContainSubstring(expected))

			overviews := system.NewZoneOverviews(
				system.ZoneResourceList{Items: []*system.ZoneResource{zone}},
				system.ZoneInsightResourceList{},
			)
			Expect(overviews.Items).To(HaveLen(1))
			Expect(overviewJSON(overviews.Items[0])).To(ContainSubstring(expected))
		},
		Entry("nil spec", nil, `"zone":{}`),
		Entry("empty spec", &system_proto.Zone{}, `"zone":{}`),
		Entry("disabled zone", &system_proto.Zone{Enabled: util_proto.Bool(false)}, `"zone":{"enabled":false}`),
	)
})
