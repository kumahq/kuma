package test

import (
	"fmt"
	"net"
)

func GetFreePort() (int, error) {
	port, err := FindFreePort("")
	return int(port), err
}

func FindFreePort(ip string) (uint32, error) {
	ln, err := net.Listen("tcp", fmt.Sprintf("%s:0", ip))
	if err != nil {
		return 0, err
	}
	if err := ln.Close(); err != nil {
		return 0, err
	}
	return uint32(ln.Addr().(*net.TCPAddr).Port), nil
}
