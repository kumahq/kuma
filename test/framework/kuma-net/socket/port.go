package socket

import (
	"math"
	"math/rand"
	"time"
)

func GenerateRandomPortsSlice(num uint, restrictedPort ...uint16) []uint16 {
	var ports []uint16

	for port := range GenerateRandomPorts(num, restrictedPort...) {
		ports = append(ports, port)
	}

	return ports
}

func GenerateRandomPorts(num uint, restrictedPort ...uint16) map[uint16]struct{} {
	rand.Seed(time.Now().UnixNano())
	randomPorts := map[uint16]struct{}{}
	restrictedPorts := map[uint16]struct{}{}

	for _, port := range restrictedPort {
		restrictedPorts[port] = struct{}{}
	}

	for len(randomPorts) < int(num) {
		// Draw a port in the range of <1,65535>
		drawn := uint16(rand.Intn(math.MaxUint16-1) + 1)

		if _, ok := restrictedPorts[drawn]; ok {
			continue
		}

		// Check if we haven't already draw this port and if our test server is not
		// exposed on currently drawn port
		if _, ok := randomPorts[drawn]; !ok {
			randomPorts[drawn] = struct{}{}
		}
	}

	return randomPorts
}
