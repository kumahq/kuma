package dns

import (
	"net"

	"github.com/Nordix/simple-ipam/pkg/ipam"
)

type IPAM interface {
	AllocateIP() (string, error)
	FreeIP(ip string) error
}

type SimpleIPAM struct {
	ipam.IPAM
}

func NewSimpleIPAM(cidr string) (IPAM, error) {
	newIPAM, err := ipam.New(cidr)
	if err != nil {
		return nil, err
	}

	return &SimpleIPAM{*newIPAM}, nil
}

func (i *SimpleIPAM) AllocateIP() (string, error) {
	ip, err := i.Allocate()
	if err != nil {
		return "", err
	}

	return ip.String(), nil
}

func (i *SimpleIPAM) FreeIP(ip string) error {
	parsedIP := net.ParseIP(ip)

	// ensure the IP is reserved before deleting it
	err := i.Reserve(parsedIP)
	if err != nil && err.Error() != "Address already allocated" {
		return err
	}

	i.Free(parsedIP)

	return nil
}
