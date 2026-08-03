package store_test

import (
	"testing"

	"github.com/kumahq/kuma/v3/pkg/test"
)

func TestSyncResourceStore(t *testing.T) {
	test.RunSpecs(t, "SyncResourceStore Delta Suite")
}
