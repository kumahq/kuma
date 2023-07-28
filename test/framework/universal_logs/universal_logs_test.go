package universal_logs_test

import (
	"path"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/types"

	"github.com/kumahq/kuma/test/framework/universal_logs"
)

func matchTempWithTimePrefix(suffix string) types.GomegaMatcher {
	return MatchRegexp(path.Join("/tmp", "[0-9]{6}_[0-9]{6}", suffix))
}

var _ = Describe("before-all and it", Ordered, func() {
	BeforeAll(func() {
		Expect(universal_logs.CurrentLogsPath("/tmp")).To(
			matchTempWithTimePrefix("/before-all-and-it"))
	})

	It("it 1", func() {})
})

var _ = Describe("before-all and context", Ordered, func() {
	BeforeAll(func() {
		Expect(universal_logs.CurrentLogsPath("/tmp")).To(
			matchTempWithTimePrefix("/before-all-and-context"))
	})

	Context("ctx 1", func() {
		It("it 1", func() {})
	})
})

var _ = Describe("before-all inside the context", Ordered, func() {
	Context("ctx 1", func() {
		BeforeAll(func() {
			Expect(universal_logs.CurrentLogsPath("/tmp")).To(
				matchTempWithTimePrefix("/before-all-inside-the-context/ctx-1"))
		})

		Context("ctx 2", func() {
			It("it 1", func() {})
		})
	})
})

var _ = Describe("before-all and describe-table", Ordered, func() {
	BeforeAll(func() {
		Expect(universal_logs.CurrentLogsPath("/tmp")).To(
			matchTempWithTimePrefix("/before-all-and-describe-table"),
		)
	})

	DescribeTable("table 1",
		func() {},
		Entry("entry 1"),
		Entry("entry 2"),
	)

	DescribeTable("table 2",
		func() {},
		Entry("entry 1"),
		Entry("entry 2"),
	)

	AfterEach(func() {
		Expect(universal_logs.CurrentLogsPath("/tmp")).To(
			matchTempWithTimePrefix("/before-all-and-describe-table/table-[0-9]{1}/entry-[0-9]{1}"),
		)
	})

	AfterAll(func() {
		Expect(universal_logs.CurrentLogsPath("/tmp")).To(
			matchTempWithTimePrefix("/before-all-and-describe-table"),
		)
	})
})
