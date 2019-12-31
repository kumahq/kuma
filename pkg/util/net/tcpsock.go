package net

import (
	"fmt"
	"net"
)

func PickTCPPort(ip string, leftPort, rightPort uint32) (actualPort uint32, err error) {
	lowestPort, highestPort := leftPort, rightPort
	if highestPort < lowestPort {
		lowestPort, highestPort = highestPort, lowestPort
	}
	// we prefer a port to remain stable over time, that's why we do sequential availability check
	// instead of random selection
	for port := lowestPort; port <= highestPort; port++ {
		if actualPort, err = ReserveTCPAddr(fmt.Sprintf("%s:%d", ip, port)); err == nil {
			return actualPort, nil
		}
	}
	return 0, err
}

func ReserveTCPAddr(address string) (uint32, error) {
	addr, err := net.ResolveTCPAddr("tcp", address)
	if err != nil {
		return 0, err
	}
	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return 0, err
	}
	defer l.Close()
	return uint32(l.Addr().(*net.TCPAddr).Port), nil
}
