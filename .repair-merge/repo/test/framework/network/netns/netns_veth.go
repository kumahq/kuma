package netns

import (
	"net"

	"github.com/vishvananda/netlink"
)

type Veth struct {
	veth      *netlink.Veth
	name      string
	peerName  string
	ipNet     *net.IPNet
	peerIPNet *net.IPNet
}

func (v *Veth) Veth() *netlink.Veth {
	return v.veth
}

func (v *Veth) PeerName() string {
	return v.peerName
}

func (v *Veth) Address() net.IP {
	return v.ipNet.IP
}

func (v *Veth) AddressCIDR() *net.IPNet {
	return v.ipNet
}

func (v *Veth) PeerAddress() net.IP {
	return v.peerIPNet.IP
}

func (v *Veth) PeerAddressCIDR() *net.IPNet {
	return v.peerIPNet
}

func (v *Veth) Name() string {
	return v.name
}
