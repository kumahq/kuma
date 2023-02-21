package udp

import (
	"fmt"
	"net"
	"runtime"

	"github.com/onsi/ginkgo/v2"

	"github.com/kumahq/kuma/test/framework/network/netns"
)

// UnsafeStartUDPServer will start TCP server in provided *netns.NesNS.
// Every initialized udp "connection" will be processed via provided callback
// functions. It was named UnsafeStartUDPServer instead of StartUDPServer
// because you have to be very cautious and remember to not spawn new goroutines
// inside provided callbacks (more info in warning below)
//
// WARNING!:
//
//	Don't spawn new goroutines inside callback functions as the goroutine inside
//	UnsafeStartUDPServer function have exclusive access to the current network
//	namespace, and you should assume, that any new goroutine will be placed
//	in a different namespace
func UnsafeStartUDPServer(
	ns *netns.NetNS,
	address string,
	processConn func(conn *net.UDPConn) error,
	callbacks ...func() error,
) (<-chan struct{}, <-chan error) {
	readyC := make(chan struct{})
	errorC := make(chan error)

	go func() {
		defer ginkgo.GinkgoRecover()
		defer close(errorC)

		runtime.LockOSThread()

		if err := ns.Set(); err != nil {
			errorC <- fmt.Errorf("cannot switch to the namespace: %s", err)
			return
		}
		defer ns.Unset() //nolint:errcheck

		addr, err := net.ResolveUDPAddr("udp", address)
		if err != nil {
			errorC <- fmt.Errorf("cannot parse address and port %q: %s", address, err)
			return
		}

		for _, callback := range callbacks {
			if err := callback(); err != nil {
				errorC <- err
				return
			}
		}

		udpConn, err := net.ListenUDP("udp", addr)
		if err != nil {
			errorC <- fmt.Errorf("cannot listen udp on address %q: %s", address, err)
			return
		}
		defer udpConn.Close()

		// At this point we are ready for accepting UDP datagrams
		close(readyC)

		for {
			if err := processConn(udpConn); err != nil {
				errorC <- err
				return
			}
		}
	}()

	return readyC, errorC
}

func ReplyWithReceivedMsg(conn *net.UDPConn) error {
	buf := make([]byte, 1024)
	n, clientAddr, err := conn.ReadFromUDP(buf)
	if err != nil {
		return fmt.Errorf("cannot read from udp: %s", err)
	}

	if _, err := conn.WriteToUDP(buf[:n], clientAddr); err != nil {
		return fmt.Errorf("cannot write to udp: %s", err)
	}

	return nil
}

func ReplyWithMsg(message string) func(conn *net.UDPConn) error {
	return func(conn *net.UDPConn) error {
		buf := make([]byte, 1024)
		_, clientAddr, err := conn.ReadFromUDP(buf)
		if err != nil {
			return fmt.Errorf("cannot read from udp: %s", err)
		}

		if _, err := conn.WriteToUDP([]byte(message), clientAddr); err != nil {
			return fmt.Errorf("cannot write to udp: %s", err)
		}

		return nil
	}
}

func ReplyWithLocalAddr(conn *net.UDPConn) error {
	_, clientAddr, err := conn.ReadFromUDP(nil)
	if err != nil {
		return fmt.Errorf("cannot read from udp: %s", err)
	}

	if _, err := conn.WriteToUDP([]byte(conn.LocalAddr().String()), clientAddr); err != nil {
		return fmt.Errorf("cannot write to udp: %s", err)
	}

	return nil
}
