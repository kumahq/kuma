package etcd_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/testcontainers/testcontainers-go"

	"github.com/kumahq/kuma/pkg/test"
	test_etcd "github.com/kumahq/kuma/pkg/test/store/etcd"
)

var c test_etcd.EtcdContainer

func TestEtcd(t *testing.T) {
	testcontainers.SkipIfProviderIsNotHealthy(t)
	BeforeSuite(func() {
		c = test_etcd.EtcdContainer{}
		Expect(c.Start()).To(Succeed())
	})
	AfterSuite(func() {
		if err := c.Stop(); err != nil {
			// Exception around delete image bug: https://github.com/moby/moby/issues/44290
			Expect(err.Error()).To(ContainSubstring("unrecognized image"))
		}
	})
	test.RunSpecs(t, "Etcd Suite")
}
