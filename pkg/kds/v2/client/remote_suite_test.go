package client_test

import (
	"testing"

	"github.com/kumahq/kuma/pkg/test"
)

func TestZoneSync(t *testing.T) {
	test.RunSpecs(t, "Zone Delta Sync Suite")
}
