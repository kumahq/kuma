package filters

import (
	"context"
	"net/http"
	"net/http/httptest"

	"github.com/emicklei/go-restful/v3"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/v3/pkg/test/resources/builders"
)

func filterRequest(rawQuery string) *restful.Request {
	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/resources?"+rawQuery, http.NoBody)
	return &restful.Request{Request: req}
}

var _ = Describe("Resource filters", func() {
	Describe("labelFilter", func() {
		DescribeTable("rejects malformed label filters",
			func(rawQuery, message string) {
				filter, err := labelFilter(filterRequest(rawQuery))

				Expect(filter).To(BeNil())
				Expect(err).To(MatchError(ContainSubstring(message)))
			},
			Entry("missing closing bracket", "filter[labels.team=red", "advanced filters are not supported"),
			Entry("empty label", "filter[labels.]=red", "label name cannot be empty"),
			Entry("nested bracket", "filter[labels.team[]=red", "name part must consist of"),
			Entry("empty value", "filter[labels.team]=", "filter value cannot be empty"),
			Entry("multiple values", "filter[labels.team]=red&filter[labels.team]=blue", "multiple filter values are not supported"),
			Entry("advanced operation", "filter[labels.team][eq]=red", "advanced filters are not supported"),
		)

		It("filters by a valid label", func() {
			filter, err := labelFilter(filterRequest("filter[labels.team]=red"))

			Expect(err).ToNot(HaveOccurred())
			Expect(filter).ToNot(BeNil())
			Expect(filter(builders.Dataplane().WithLabels(map[string]string{"team": "red"}).Build())).To(BeTrue())
			Expect(filter(builders.Dataplane().WithLabels(map[string]string{"team": "blue"}).Build())).To(BeFalse())
		})

		It("leaves status filtering to the endpoint", func() {
			filter, err := labelFilter(filterRequest("filter[status]=online"))

			Expect(err).ToNot(HaveOccurred())
			Expect(filter).To(BeNil())
		})
	})
})
