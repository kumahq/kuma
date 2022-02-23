package xds_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/pkg/util/xds"
)

var _ = Describe("Metric sanitizer", func() {
	It("should sanitize metrics", func() {
		// given
		metric := "some metric with chars :/_-0123{version=3.0}"

		// when
		sanitized := xds.SanitizeMetric(metric)

		// then
		Expect(sanitized).To(Equal("some_metric_with_chars____-0123_version_3_0_"))
	})
})
