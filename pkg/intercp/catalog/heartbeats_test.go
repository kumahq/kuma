package catalog_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/pkg/intercp/catalog"
)

var _ = Describe("Heartbeats", func() {
	var heartbeats *catalog.Heartbeats

	BeforeEach(func() {
		heartbeats = catalog.NewHeartbeats()
	})

	It("should remove heartbeats on collection", func() {
		// given
		instance := catalog.Instance{
			Id: "instance-1",
		}
		heartbeats.Add(instance)

		// when
		instances := heartbeats.ResetAndCollect()

		// then
		Expect(instances).To(HaveLen(1))
		Expect(instances[0]).To(Equal(instance))

		// when
		instances = heartbeats.ResetAndCollect()

		// then
		Expect(instances).To(BeEmpty())
	})
})
