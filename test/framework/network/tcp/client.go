package tcp

import (
	"fmt"
	"net"
)

func DialIPWithPortAndGetReply(ip net.IP, port uint16) (string, error) {
	return DialTCPAddrAndGetReply(&net.TCPAddr{
		IP:   ip,
		Port: int(port),
	})
}

func DialTCPAddrAndGetReply(address *net.TCPAddr) (string, error) {
	conn, err := net.DialTCP("tcp", nil, address)
	if err != nil {
		return "", fmt.Errorf("cannot dial provided address: %s", err)
	}
	defer conn.Close()

	return readBytes(conn)
}

func readBytes(conn *net.TCPConn) (string, error) {
	buff := make([]byte, 1024)
	n, err := conn.Read(buff)
	if err != nil {
		return "", fmt.Errorf("cannot read from the connection: %s", err)
	}

	if err := conn.Close(); err != nil {
		return "", fmt.Errorf("cannot close the connection: %s", err)
	}

	return string(buff[:n]), nil
}
