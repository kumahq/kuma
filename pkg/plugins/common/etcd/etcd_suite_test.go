package etcd_test

import (
	"testing"

	"github.com/kumahq/kuma/pkg/test"
)

func TestEtcd(t *testing.T) {
	test.RunSpecs(t, "Etcd Suite")
}
