package catalog_test

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/pkg/core/resources/manager"
	"github.com/kumahq/kuma/pkg/intercp/catalog"
	"github.com/kumahq/kuma/pkg/plugins/resources/memory"
)

var _ = Describe("Catalog", func() {

	var c catalog.Catalog

	BeforeEach(func() {
		store := memory.NewStore()
		resManager := manager.NewResourceManager(store)
		c = catalog.NewConfigCatalog(resManager)
	})

	Context("Replace", func() {
		instances := []catalog.Instance{
			{
				Id:     "instance-1",
				Leader: true,
			},
		}

		BeforeEach(func() {
			_, err := c.Replace(context.Background(), instances)
			Expect(err).ToNot(HaveOccurred())
		})

		It("should replace existing instances", func() {
			// when
			instances := []catalog.Instance{
				{
					Id:     "instance-1",
					Leader: false,
				},
				{
					Id:     "instance-2",
					Leader: true,
				},
			}
			updated, err := c.Replace(context.Background(), instances)

			// then
			Expect(updated).To(BeTrue())
			Expect(err).ToNot(HaveOccurred())

			readInstances, err := c.Instances(context.Background())
			Expect(readInstances).To(Equal(instances))
			Expect(err).ToNot(HaveOccurred())
		})

		It("should return false if replace did not replace instances", func() {
			// when
			updated, err := c.Replace(context.Background(), instances)

			// then
			Expect(updated).To(BeFalse())
			Expect(err).ToNot(HaveOccurred())
		})
	})

	Context("Leader", func() {
		It("should return a leader if there is a leader in the list", func() {
			// given
			instances := []catalog.Instance{
				{
					Id:     "instance-1",
					Leader: false,
				},
				{
					Id:     "instance-2",
					Leader: true,
				},
			}
			_, err := c.Replace(context.Background(), instances)
			Expect(err).ToNot(HaveOccurred())

			// when
			leader, err := catalog.Leader(context.Background(), c)

			// then
			Expect(err).ToNot(HaveOccurred())
			Expect(leader.Id).To(Equal("instance-2"))
		})

		It("should return an error if there is no leader", func() {
			// given
			instances := []catalog.Instance{
				{
					Id:     "instance-1",
					Leader: false,
				},
			}
			_, err := c.Replace(context.Background(), instances)
			Expect(err).ToNot(HaveOccurred())

			// when
			_, err = catalog.Leader(context.Background(), c)

			// then
			Expect(err).To(Equal(catalog.ErrNoLeader))
		})
	})

	It("should return empty instances if the catalog was never updated", func() {
		// when
		instances, err := c.Instances(context.Background())

		// then
		Expect(err).ToNot(HaveOccurred())
		Expect(instances).To(HaveLen(0))
	})
})
