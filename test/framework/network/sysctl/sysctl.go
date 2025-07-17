package sysctl

import (
	"fmt"
	"os"
	"strconv"
)

func set(path, value string) error {
	return os.WriteFile(path, []byte(value+"\n"), 0o600)
}

// SetLocalPortRange will change the range of local ports available for udp/tcp
// sockets
//
// IPv6 stack is using this setting as well
// ref. https://tldp.org/HOWTO/Linux+IPv6-HOWTO/ch11s03.html
func SetLocalPortRange(minimum, maximum uint16) func() error {
	return func() error {
		return set("/proc/sys/net/ipv4/ip_local_port_range", fmt.Sprintf("%d\t%d", minimum, maximum))
	}
}

func SetUnprivilegedPortStart(value uint16) func() error {
	return func() error {
		return set("/proc/sys/net/ipv4/ip_unprivileged_port_start", strconv.Itoa(int(value)))
	}
}
