package vips

import (
	"net"

	"github.com/Nordix/simple-ipam/pkg/ipam"
)

// GlobalView keeps a list of all hostname/ips and add the possibility to allocate new ips
type GlobalView struct {
	ipam         *ipam.IPAM
	ipToHostname map[string]HostnameEntry
	hostnameToIp map[HostnameEntry]string
}

// Reserve add an ip/host to the list of reserved ips (useful when loading an existing view).
func (g *GlobalView) Reserve(hostname HostnameEntry, ip string) error {
	err := g.ipam.Reserve(net.ParseIP(ip))
	if err != nil {
		return err
	}
	g.ipToHostname[ip] = hostname
	g.hostnameToIp[hostname] = ip
	return nil
}

// Allocate assign an ip to a host
func (g *GlobalView) Allocate(hostname HostnameEntry) (string, error) {
	ip := g.hostnameToIp[hostname]
	if ip != "" {
		return ip, nil
	}
	netIp, err := g.ipam.Allocate()
	if err != nil {
		return "", err
	}
	ip = netIp.String()
	g.ipToHostname[ip] = hostname
	g.hostnameToIp[hostname] = ip
	return ip, nil
}

func (g *GlobalView) ToVIPMap() map[HostnameEntry]string {
	return g.hostnameToIp
}

func NewGlobalView(cidr string) (*GlobalView, error) {
	newIPAM, err := ipam.New(cidr)
	if err != nil {
		return nil, err
	}

	return &GlobalView{
		hostnameToIp: map[HostnameEntry]string{},
		ipToHostname: map[string]HostnameEntry{},
		ipam:         newIPAM,
	}, nil
}
