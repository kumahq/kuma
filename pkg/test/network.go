package test

import (
	"context"
	"fmt"
	"net"
	"sync"

	"github.com/pkg/errors"
)

// The kernel is free to hand out the same ephemeral port again once the socket
// used to find it is closed, which makes two consecutive calls return the same
// port every few thousand times on Linux. Tests that pick one port per server
// then configure two servers with the same port, so ports already handed out in
// this process are not returned again.
var handedOutPorts = struct {
	sync.Mutex
	ports map[uint32]struct{}
}{ports: map[uint32]struct{}{}}

func GetFreePort() (int, error) {
	port, err := FindFreePort("")
	return int(port), err
}

func FindFreePort(ip string) (uint32, error) {
	for range 10 {
		port, err := findFreePort(ip)
		if err != nil {
			return 0, err
		}
		handedOutPorts.Lock()
		_, handedOut := handedOutPorts.ports[port]
		if !handedOut {
			handedOutPorts.ports[port] = struct{}{}
		}
		handedOutPorts.Unlock()
		if !handedOut {
			return port, nil
		}
	}
	return 0, errors.New("could not find a free port that was not handed out before")
}

func findFreePort(ip string) (uint32, error) {
	ln, err := (&net.ListenConfig{}).Listen(context.Background(), "tcp", fmt.Sprintf("%s:0", ip))
	if err != nil {
		return 0, err
	}
	if err := ln.Close(); err != nil {
		return 0, err
	}
	return uint32(ln.Addr().(*net.TCPAddr).Port), nil
}
