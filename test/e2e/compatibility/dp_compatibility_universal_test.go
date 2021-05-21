package compatibility_test

import (
	. "github.com/onsi/ginkgo"

	"github.com/kumahq/kuma/test/e2e/compatibility"
)

var _ = Describe("Test Universal Compatibility", compatibility.UniversalCompatibility)
