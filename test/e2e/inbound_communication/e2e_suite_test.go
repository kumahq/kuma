package inbound_communication_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"

	"github.com/kumahq/kuma/pkg/test"
	"github.com/kumahq/kuma/test/e2e/inbound_communication"
)

func TestE2E(t *testing.T) {
	test.RunSpecs(t, "E2E Server Bind")
}

var _ = Describe("Test Application Server Bind on Multizone", inbound_communication.ServerBind, Ordered)
