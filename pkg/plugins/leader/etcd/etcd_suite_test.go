package etcd_test

import (
	"testing"

	"github.com/testcontainers/testcontainers-go"

	"github.com/kumahq/kuma/pkg/test"
)

func TestEtcd(t *testing.T) {
	testcontainers.SkipIfProviderIsNotHealthy(t)
	test.RunSpecs(t, "Etcd Leader Suite")
}
