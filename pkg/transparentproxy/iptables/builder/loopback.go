package builder

import (
	"fmt"
	"net"
)

func getLoopback() (*net.Interface, error) {
	interfaces, err := net.Interfaces()
	if err != nil {
		return nil, fmt.Errorf("listig network interfaces failed: %s", err)
	}

	for _, iface := range interfaces {
		if iface.Flags&net.FlagLoopback != 0 {
			return &iface, nil
		}
	}

	return nil, fmt.Errorf("it appears there is no loopback interface")
}
