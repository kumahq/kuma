package store_test

import (
	"testing"

	"github.com/kumahq/kuma/v2/pkg/test"
)

func TestSyncResourceStore(t *testing.T) {
	test.RunSpecs(t, "SyncResourceStore Delta Suite")
}
