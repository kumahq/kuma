package matchers

import (
	"math"

	"github.com/onsi/gomega"
	"github.com/onsi/gomega/types"
)

func BeValidPort() types.GomegaMatcher {
	return gomega.SatisfyAll(
		gomega.BeNumerically(">=", uint16(1)),
		gomega.BeNumerically("<=", math.MaxUint16))
}
