package types

import (
	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
)

type Outbound struct {
	// LegacyOutbound is an old way to define outbounds using 'kuma.io/service' tag
	LegacyOutbound *mesh_proto.Dataplane_Networking_Outbound

	Address  string
	Port     uint32
	Resource *core_model.TypedResourceIdentifier
}

func (o *Outbound) GetAddress() string {
	if o.LegacyOutbound != nil {
		return o.LegacyOutbound.Address
	}
	return o.Address
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
