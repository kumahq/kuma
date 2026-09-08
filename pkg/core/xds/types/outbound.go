package types

import (
	"github.com/kumahq/kuma/v3/pkg/core/kri"
)

type Outbound struct {
	Address  string
	Port     uint32
	Resource kri.Identifier
}

// AssociatedServiceResource returns KRI to MeshService, MeshExternalService or
// MeshMultiZoneService this outbound points at, and true when it is set.
func (o *Outbound) AssociatedServiceResource() (kri.Identifier, bool) {
	return o.Resource, !o.Resource.IsEmpty()
}

func (o *Outbound) GetAddress() string {
	return o.GetAddressWithFallback("")
}

// GetAddressWithFallback returns the address of the outbound, or the fallback
// value when it is empty.
func (o *Outbound) GetAddressWithFallback(fallback string) string {
	if o.Address != "" {
		return o.Address
	}
	return fallback
}

func (o *Outbound) GetPort() uint32 {
	return o.Port
}

type Outbounds []*Outbound
