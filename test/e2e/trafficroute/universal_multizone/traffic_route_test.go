package universal_multizone_test

import (
	. "github.com/onsi/ginkgo"

	"github.com/kumahq/kuma/test/e2e/trafficroute/universal_multizone"
)

var _ = Describe("Test Multizone Universal deployment", universal_multizone.KumaMultizone)
