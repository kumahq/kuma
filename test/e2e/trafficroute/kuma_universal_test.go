package trafficroute_test

import (
	"github.com/kumahq/kuma/test/e2e/trafficroute"

	. "github.com/onsi/ginkgo"
)

var _ = Describe("Test Standalone Universal deployment", trafficroute.KumaStandalone)

var _ = Describe("Test Multizone Universal deployment", trafficroute.KumaMultizone)
