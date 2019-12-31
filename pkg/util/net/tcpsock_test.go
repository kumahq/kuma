package net_test

import (
	"fmt"
	"net"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	. "github.com/Kong/kuma/pkg/util/net"

	"github.com/Kong/kuma/pkg/test"
)

var _ = Describe("ReserveTCPAddr()", func() {
	It("should successfully reserve a free TCP address (ip + port)", func() {
		// given
		loopback := "127.0.0.1"

		// setup
		freePort, err := test.FindFreePort(loopback)
		Expect(err).ToNot(HaveOccurred())
		// and
		address := fmt.Sprintf("%s:%d", loopback, freePort)

		// when
		actualPort, err := ReserveTCPAddr(address)
		// then
		Expect(err).ToNot(HaveOccurred())
		// and
		Expect(actualPort).To(Equal(freePort))
	})

	It("should fail to reserve a TCP address already in use (ip + port)", func() {
		// given
		loopback := "127.0.0.1"

		// setup
		freePort, err := test.FindFreePort(loopback)
		Expect(err).ToNot(HaveOccurred())
		// and
		address := fmt.Sprintf("%s:%d", loopback, freePort)

		By("simulating another Envoy instance that already uses this port")
		// when
		l, err := net.Listen("tcp", address)
		// then
		Expect(err).ToNot(HaveOccurred())
		// and
		defer l.Close()

		// when
		actualPort, err := ReserveTCPAddr(address)
		// then
		Expect(err.Error()).To(ContainSubstring(`bind: address already in use`))
		// and
		Expect(actualPort).To(Equal(uint32(0)))
	})
})

var _ = Describe("PickTCPPort()", func() {

	It("should be able to pick the 1st port in the range", func() {
		// given
		loopback := "127.0.0.1"

		// setup
		freePort, err := test.FindFreePort(loopback)
		Expect(err).ToNot(HaveOccurred())

		// when
		actualPort, err := PickTCPPort(loopback, freePort, freePort)
		// then
		Expect(err).ToNot(HaveOccurred())
		// and
		Expect(actualPort).To(Equal(freePort))
	})

	Describe("should be able to pick the Nth port in the range", func() {

		// given
		loopback := "127.0.0.1"

		findFreePortRange := func(n uint32) (lowestPort uint32, highestPort uint32) {
			Expect(n).To(BeNumerically(">", 0))
		attempts:
			for a := 0; a < 65535; a++ {
				// first port in a range
				freePort, err := test.FindFreePort(loopback)
				Expect(err).ToNot(HaveOccurred())

				// next n-1 ports in that range
				for i := uint32(1); i < n; i++ {
					address := fmt.Sprintf("%s:%d", loopback, freePort+i)
					if _, err := ReserveTCPAddr(address); err != nil {
						continue attempts
					}
				}

				return freePort, freePort + n - 1
			}
			Fail(fmt.Sprintf(`unable to find "%d" free ports in a row`, n))
			return
		}

		type testCase struct {
			n uint32
		}

		testSet := func(n uint32) []TableEntry {
			cases := make([]TableEntry, 0, n)
			for i := uint32(2); i <= n; i++ {
				cases = append(cases, Entry(fmt.Sprintf("%d", i), testCase{n: i}))
			}
			return cases
		}

		DescribeTable("should be able to pick the Nth port in the range",
			func(given testCase) {
				By("finding N consecutive free ports in a row")
				lowestPort, highestPort := findFreePortRange(given.n)

				By("simulating another Envoy instances using first N-1 ports")
				for i := uint32(0); i < given.n-1; i++ {
					// given
					address := fmt.Sprintf("%s:%d", loopback, lowestPort+i)
					// when
					l, err := net.Listen("tcp", address)
					// then
					Expect(err).ToNot(HaveOccurred())
					// and
					defer l.Close()
				}

				// when
				actualPort, err := PickTCPPort(loopback, lowestPort, highestPort)
				// then
				Expect(err).ToNot(HaveOccurred())
				// and
				Expect(actualPort).To(Equal(highestPort))
			},
			testSet(10)...,
		)
	})

	It("should fail to pick a free port when all ports in the range are in use", func() {
		// given
		loopback := "127.0.0.1"

		// setup
		freePort, err := test.FindFreePort(loopback)
		Expect(err).ToNot(HaveOccurred())
		// and
		address := fmt.Sprintf("%s:%d", loopback, freePort)

		By("simulating another Envoy instance that already uses this port")
		// when
		l, err := net.Listen("tcp", address)
		// then
		Expect(err).ToNot(HaveOccurred())
		// and
		defer l.Close()

		// when
		actualPort, err := PickTCPPort(loopback, freePort, freePort)
		// then
		Expect(err.Error()).To(ContainSubstring(`bind: address already in use`))
		// and
		Expect(actualPort).To(Equal(uint32(0)))
	})

	It("should be able to pick a random port", func() {
		// given
		loopback := "127.0.0.1"

		// when
		actualPort, err := PickTCPPort(loopback, 0, 0)
		// then
		Expect(err).ToNot(HaveOccurred())
		// and
		Expect(actualPort).ToNot(Equal(uint32(0)))
	})

	It("should re-order port range bounds if necessary", func() {
		// given
		loopback := "127.0.0.1"

		// when
		actualPort, err := PickTCPPort(loopback, 1, 0)
		// then
		Expect(err).ToNot(HaveOccurred())
		// and
		Expect(actualPort).ToNot(Equal(uint32(0)))
	})
})
