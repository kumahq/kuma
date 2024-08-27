//go:build !windows

package probes

import (
	"net"
	"syscall"
)

// copied from https://github.com/kubernetes/kubernetes/blob/v1.27.0-alpha.1/pkg/probe/dialer_others.go#L27
// createProbeDialer returns a dialer optimized for probes to avoid lingering sockets on TIME-WAIT state.
// The dialer reduces the TIME-WAIT period to 1 seconds instead of the OS default of 60 seconds.
// Using 1 second instead of 0 because SO_LINGER socket option to 0 causes pending data to be
// discarded and the connection to be aborted with an RST rather than for the pending data to be
// transmitted and the connection closed cleanly with a FIN.
// Ref: https://issues.k8s.io/89898
func createProbeDialer(isIPv6 bool) *net.Dialer {
	dialer := &net.Dialer{
		Control: func(network, address string, c syscall.RawConn) error {
			return c.Control(func(fd uintptr) {
				_ = syscall.SetsockoptLinger(int(fd), syscall.SOL_SOCKET, syscall.SO_LINGER, &syscall.Linger{Onoff: 1, Linger: 1})
			})
		},
	}
	dialer.LocalAddr = LocalAddrIPv4
	if isIPv6 {
		dialer.LocalAddr = LocalAddrIPv6
	}
	return dialer
}
