package utils

import (
	"fmt"
	"math/rand"
	"net"
	"strings"
	"sync"
)

func ShellEscape(arg string) string {
	return fmt.Sprintf("'%s'", strings.ReplaceAll(arg, "'", "\\'"))
}

var usedPorts map[int]struct{}
var mutex sync.Mutex

func init() {
	usedPorts = map[int]struct{}{}
}

func GetFreePort() (int, error) {
	mutex.Lock()
	defer mutex.Unlock()
	for {
		port := rand.Int()%50000 + 10000
		if _, ok := usedPorts[port]; ok {
			continue
		}
		address, err := net.ResolveTCPAddr("tcp", fmt.Sprintf(":%d", port))
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
		usedPorts[port] = struct{}{}
		return port, nil
	}
}
