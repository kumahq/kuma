package universal_standalone_test

import (
	"github.com/kumahq/kuma/test/e2e/trafficroute/universal_standalone"
	. "github.com/onsi/ginkgo"
)

var _ = Describe("Test Standalone Universal deployment", universal_standalone.KumaStandalone)
