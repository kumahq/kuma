package types

import (
	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/kri"
)

type Outbound struct {
	// LegacyOutbound is an old way to define outbounds using 'kuma.io/service' tag
	LegacyOutbound *mesh_proto.Dataplane_Networking_Outbound

	Address  string
	Port     uint32
	Resource kri.Identifier
}

// AssociatedServiceResource
//   - if the outbound supports new service resources,
//     then it returns KRI to MeshService, MeshExternalService, MeshMultiZoneService and true
//   - if the outbound is defined using the old way with 'kuma.io/service' tag,
//     then it returns empty kri.Identifier and false
func (o *Outbound) AssociatedServiceResource() (kri.Identifier, bool) {
	return o.Resource, !o.Resource.IsEmpty()
}

func (o *Outbound) GetAddress() string {
	return o.GetAddressWithFallback("")
}

// GetAddressWithFallback returns the address from LegacyOutbound if set,
// otherwise from Address. If both are empty, it returns the fallback value.
func (o *Outbound) GetAddressWithFallback(fallback string) string {
	switch {
	case o.LegacyOutbound != nil && o.LegacyOutbound.Address != "":
		return o.LegacyOutbound.Address
	case o.Address != "":
		return o.Address
	default:
		return fallback
	}
}

// TagsOrNil returns tags if Outbound is defined using 'kuma.io/service' tag and so LegacyOutbound field is set.
// Otherwise, it returns nil.
func (o *Outbound) TagsOrNil() map[string]string {
	if o.LegacyOutbound != nil {
		return o.LegacyOutbound.Tags
	}
	return nil
}

func (o *Outbound) GetPort() uint32 {
	if o.LegacyOutbound != nil {
		return o.LegacyOutbound.Port
	}
	return o.Port
}

type Outbounds []*Outbound

func (os Outbounds) Filter(predicates ...func(o *Outbound) bool) Outbounds {
	var result []*Outbound
	for _, outbound := range os {
		add := true
		for _, p := range predicates {
			if !p(outbound) {
				add = false
			}
		}
		if add {
			result = append(result, outbound)
		}
	}
	return result
}

func NonBackendRefFilter(o *Outbound) bool {
	return o.LegacyOutbound != nil && o.LegacyOutbound.BackendRef == nil
}
