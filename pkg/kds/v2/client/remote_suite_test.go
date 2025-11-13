package client_test

import (
	"testing"

	"github.com/kumahq/kuma/v2/pkg/test"
)

func TestZoneSync(t *testing.T) {
	test.RunSpecs(t, "Zone Delta Sync Suite")
}
