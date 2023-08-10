package etcd_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/pkg/core/resources/store"
	core_metrics "github.com/kumahq/kuma/pkg/metrics"
	"github.com/kumahq/kuma/pkg/plugins/resources/etcd"
	test_store "github.com/kumahq/kuma/pkg/test/store"
)

var _ = Describe("StoreTemplate", func() {
	createStore := func(prefix string) func() store.ResourceStore {
		return func() store.ResourceStore {
			cfg, err := c.Config()
			Expect(err).ToNot(HaveOccurred())

			etcdMetrics, err := core_metrics.NewMetrics("etcd")
			Expect(err).ToNot(HaveOccurred())

			store, err := etcd.NewEtcdStore(prefix, etcdMetrics, cfg)
			Expect(err).ToNot(HaveOccurred())
			return store
		}
	}

	test_store.ExecuteStoreTests(createStore("test"), "etcd")
	test_store.ExecuteOwnerTests(createStore("test"), "etcd")
})
