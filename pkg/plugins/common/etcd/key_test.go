package etcd_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/pkg/plugins/common/etcd"
)

var _ = Describe("Key", func() {
	Context("etcd resource key", func() {
		It("create etcd key", func() {
			key := etcd.NewEtcdResourcedKey("test", "type", "mesh", "name").String()
			Expect(key).To(Equal("/test/resource/type/mesh/name"))
		})

		It("with etcd key", func() {
			etcdKey, err := etcd.WithEtcdKey("/test/resource/type/mesh/name")
			Expect(err).ToNot(HaveOccurred())
			Expect(etcdKey).To(Equal(etcd.NewEtcdResourcedKey("test", "type", "mesh", "name")))
		})

		It("with prefix key", func() {
			key := etcd.NewEtcdResourcedKey("test", "type", "", "").Prefix()
			Expect(key).To(Equal("/test/resource/type"))

			key = etcd.NewEtcdResourcedKey("test", "type", "mesh", "").Prefix()
			Expect(key).To(Equal("/test/resource/type/mesh"))

			key = etcd.NewEtcdResourcedKey("test", "type", "mesh", "name").Prefix()
			Expect(key).To(Equal("/test/resource/type/mesh/name"))
		})
	})

	Context("etcd resource key", func() {
		It("create index key", func() {
			key := etcd.NewEtcdIndexKey("test", &etcd.Key{
				Typ:  "type",
				Mesh: "mesh",
				Name: "name",
			}, &etcd.Key{
				Typ:  "ownertype",
				Mesh: "ownermesh",
				Name: "ownername",
			}).String()

			Expect(key).To(Equal("/test/index/ownertype/ownermesh/ownername/type/mesh/name"))
		})
	})
})
