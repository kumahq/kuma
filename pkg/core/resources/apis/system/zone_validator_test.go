package system_test

import (
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/v3/pkg/core/resources/apis/system"
	test_model "github.com/kumahq/kuma/v3/pkg/test/resources/model"
)

var _ = Describe("Zone", func() {
	DescribeTable("should validate that name conforms to RFC 1035",
		func(name string, expectedViolation string) {
			// given
			zone := system.NewZoneResource()
			zone.SetMeta(&test_model.ResourceMeta{Name: name})

			// when
			err := zone.Validate()

			// then
			if expectedViolation == "" {
				Expect(err).ToNot(HaveOccurred())
			} else {
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring(expectedViolation))
			}
		},
		Entry("valid name", "east-1", ""),
		Entry("name with a dot", "east.1", "name: a DNS-1035 label must consist of lower case alphanumeric characters"),
		Entry("name starting with a digit", "1east", "name: a DNS-1035 label must consist of lower case alphanumeric characters"),
		Entry("name with an underscore", "east_1", "name: a DNS-1035 label must consist of lower case alphanumeric characters"),
		Entry("name with an uppercase letter", "East", "name: a DNS-1035 label must consist of lower case alphanumeric characters"),
		Entry("name of 63 characters", strings.Repeat("e", 63), ""),
		Entry("name of 64 characters", strings.Repeat("e", 64), "name: must be no more than 63 characters"),
	)
})
