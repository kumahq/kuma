package socket_test

import (
	"math"
	"math/rand"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "github.com/kumahq/kuma/pkg/test/matchers"
	"github.com/kumahq/kuma/test/framework/network/socket"
)

var _ = Describe("GenerateRandomPorts", func() {
	It("should generate random ports when no restricted ports provided", func() {
		// #nosec G404 -- this is just a test
		r := rand.New(rand.NewSource(GinkgoRandomSeed()))

		for i := 0; i < 10; i++ {
			r := r.Intn(30) + 5
			ports := socket.GenerateRandomPorts(uint(r))

			Expect(ports).To(HaveLen(r))
			Expect(ports).To(AllKeys(BeValidPort()))
		}
	})

	It("should generate random ports without restricted ports", func() {
		// We will generate almost all possible ports, without 6 (5 provided by us
		// and 1 for '0'), which then we'll check if are not present in the generated
		// list
		num := math.MaxUint16 - 6
		restrictedPorts := []uint16{22, 80, 443, 8080, 8443}

		ports := socket.GenerateRandomPorts(uint(num), restrictedPorts...)

		Expect(ports).To(HaveLen(num))
		Expect(ports).To(AllKeys(Not(BeElementOf(restrictedPorts))))
	})
})
