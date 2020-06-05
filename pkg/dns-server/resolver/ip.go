package resolver

import (
	"net"

	"github.com/Nordix/simple-ipam/pkg/ipam"
)

type IPAM struct {
	ipam.IPAM
}

func (d *SimpleDNSResolver) initIPAM(cidr string) error {
	ipam, err := ipam.New(cidr)
	if err != nil {
		return err
	}

	d.ipam = &IPAM{*ipam}
	return nil
}

func (d *SimpleDNSResolver) allocateIP() (string, error) {
	ip, err := d.ipam.Allocate()
	if err != nil {
		return "", err
	}
	return ip.String(), nil
}

func (d *SimpleDNSResolver) freeIP(ip string) error {
	parsedIP := net.ParseIP(ip)
	// ensure the IP is reserved before deleting it
	err := d.ipam.Reserve(parsedIP)
	if err != nil {
		if err.Error() != "Address already allocated" {
			return err
		}
	}
	d.ipam.Free(parsedIP)
	return nil
}
