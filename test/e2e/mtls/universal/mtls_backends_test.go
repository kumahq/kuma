package universal_test

import (
	. "github.com/onsi/ginkgo"

	"github.com/kumahq/kuma/test/e2e/mtls/universal"
)

var _ = Describe("mTLS on Universal", universal.MTLSUniversal)
