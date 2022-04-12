package utils

import (
	"fmt"
	"math/rand"
	"net"
	"strings"
)

func ShellEscape(arg string) string {
	return fmt.Sprintf("'%s'", strings.ReplaceAll(arg, "'", "\\'"))
}

func GetFreePort() (int, error) {
	for {
		port := rand.Int()%50000 + 10000
		address, err := net.ResolveTCPAddr("tcp", fmt.Sprintf("localhost:%d", port))
		if err != nil {
			return 0, err
		}
		listener, err := net.ListenTCP("tcp", address)
		if err != nil {
			continue
		}
		if err := listener.Close(); err != nil {
			return 0, err
		}
		return port, nil
	}
}
