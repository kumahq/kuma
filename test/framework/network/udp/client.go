package udp

import (
	"fmt"
	"net"
	"sync"
)

type ResultMap struct {
	results map[string]uint
	mutex   sync.RWMutex
}

func NewResultMap() *ResultMap {
	return &ResultMap{
		results: map[string]uint{},
	}
}

func (m *ResultMap) Increase(key string) {
	m.mutex.Lock()
	m.results[key]++
	m.mutex.Unlock()
}

func (m *ResultMap) GetFinalResults() map[string]uint {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	results := map[string]uint{}

	for key, value := range m.results {
		results[key] = value
	}

	return results
}

// DialUDPAddrWithHelloMsgAndGetReply will open a UDP socket with provided
// UDP address and send there a helloMsg (fmt.Stringer), and block goroutine
// waiting for the message back, which will be returned as a string
func DialUDPAddrWithHelloMsgAndGetReply(
	address *net.UDPAddr,
	helloMsg fmt.Stringer,
) (string, error) {
	socket, err := net.DialUDP("udp", nil, address)
	if err != nil {
		return "", fmt.Errorf("cannot dial provided address: %s", err)
	}
	defer socket.Close()

	if _, err = socket.Write([]byte(helloMsg.String())); err != nil {
		return "", fmt.Errorf("cannot send hello message %q: %s", helloMsg, err)
	}

	buf := make([]byte, 1024)
	n, err := socket.Read(buf)
	if err != nil {
		return "", fmt.Errorf("cannot read replied message: %s", err)
	}

	return string(buf[:n]), nil
}

func DialAddrAndIncreaseResultMap(address string, resultMap *ResultMap) error {
	udpAddr, err := net.ResolveUDPAddr("udp", address)
	if err != nil {
		return fmt.Errorf("cannot resolve UDPAddr from provided address (%s): %s", address, err)
	}

	socket, err := net.DialUDP("udp", nil, udpAddr)
	if err != nil {
		return fmt.Errorf("cannot dial provided address (%s): %s", address, err)
	}
	defer socket.Close()

	if _, err := socket.Write(nil); err != nil {
		return fmt.Errorf("cannot send hello message: %s", err)
	}

	buf := make([]byte, 1024)
	n, err := socket.Read(buf)
	if err != nil {
		return fmt.Errorf("cannot read replied message: %s", err)
	}

	resultMap.Increase(string(buf[:n]))

	return nil
}
