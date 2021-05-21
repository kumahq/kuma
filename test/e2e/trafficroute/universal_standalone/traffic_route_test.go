package universal_standalone_test

import (
	. "github.com/onsi/ginkgo"

	"github.com/kumahq/kuma/test/e2e/trafficroute/universal_standalone"
)

var _ = Describe("Test Standalone Universal deployment", universal_standalone.KumaStandalone)
