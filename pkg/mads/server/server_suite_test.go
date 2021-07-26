package server_test

import (
	"testing"

	"github.com/kumahq/kuma/pkg/test"
)

func TestMADSServer(t *testing.T) {
	test.RunSpecs(t, "MADS Server Suite")
}
