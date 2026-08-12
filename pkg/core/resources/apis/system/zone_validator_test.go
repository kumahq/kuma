package system_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/v3/pkg/core/resources/apis/system"
	test_model "github.com/kumahq/kuma/v3/pkg/test/resources/model"
)

var _ = Describe("Zone", func() {
	DescribeTable("should validate that name conforms to RFC 1035",
		func(name string, expectedViolation bool) {
			// given
			zone := system.NewZoneResource()
			zone.SetMeta(&test_model.ResourceMeta{Name: name})

			// when
			err := zone.Validate()

			// then
			if expectedViolation {
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("name: a DNS-1035 label must consist of lower case alphanumeric characters"))
			} else {
				Expect(err).ToNot(HaveOccurred())
			}
		},
		Entry("valid name", "east-1", false),
		Entry("name with a dot", "east.1", true),
		Entry("name starting with a digit", "1east", true),
		Entry("name with an underscore", "east_1", true),
		Entry("name with an uppercase letter", "East", true),
	)
})
