package v1alpha1_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	zone_api "github.com/kumahq/kuma/v3/pkg/core/resources/apis/zone/api/v1alpha1"
	zoneinsight_api "github.com/kumahq/kuma/v3/pkg/core/resources/apis/zoneinsight/api/v1alpha1"
	core_model "github.com/kumahq/kuma/v3/pkg/core/resources/model"
	"github.com/kumahq/kuma/v3/pkg/core/resources/model/rest"
	"github.com/kumahq/kuma/v3/pkg/test/resources/model"
)

var _ = Describe("ZoneOverview", func() {
	overviewJSON := func(overview core_model.Resource) string {
		GinkgoHelper()
		bytes, err := rest.From.Resource(overview).(interface{ MarshalJSON() ([]byte, error) }).MarshalJSON()
		Expect(err).ToNot(HaveOccurred())
		return string(bytes)
	}

	DescribeTable("should always render a zone, whatever the Zone spec is",
		func(spec *zone_api.Zone, expected string) {
			zone := &zone_api.ZoneResource{
				Meta: &model.ResourceMeta{Name: "zone-1"},
				Spec: spec,
			}

			overview := zone_api.NewZoneOverviewResource()
			Expect(overview.SetOverviewSpec(zone, nil)).To(Succeed())
			Expect(overviewJSON(overview)).To(ContainSubstring(expected))

			overviews := zone_api.NewZoneOverviews(
				zone_api.ZoneResourceList{Items: []*zone_api.ZoneResource{zone}},
				zoneinsight_api.ZoneInsightResourceList{},
			)
			Expect(overviews.Items).To(HaveLen(1))
			Expect(overviewJSON(overviews.Items[0])).To(ContainSubstring(expected))
		},
		Entry("nil spec", nil, `"zone":{}`),
		Entry("empty spec", &zone_api.Zone{}, `"zone":{}`),
		Entry("disabled zone", &zone_api.Zone{Enabled: new(bool)}, `"zone":{"enabled":false}`),
	)
})
