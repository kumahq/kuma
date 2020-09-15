package v1alpha1

import (
	"github.com/pkg/errors"
)

func (n *ExternalService_Networking) GetInboundInterface(service string) (*InboundInterface, error) {
	for _, inbound := range n.Inbound {
		if inbound.Tags[ServiceTag] != service {
			continue
		}
		iface := n.ToInboundInterface(inbound)
		return &iface, nil
	}
	return nil, errors.Errorf("Dataplane has no Inbound Interface for service %q", service)
}

func (n *ExternalService_Networking) GetInboundInterfaces() ([]InboundInterface, error) {
	if n == nil {
		return nil, nil
	}
	ifaces := make([]InboundInterface, len(n.Inbound))
	for i, inbound := range n.Inbound {
		ifaces[i] = n.ToInboundInterface(inbound)
	}
	return ifaces, nil
}

func (n *ExternalService_Networking) ToInboundInterface(inbound *ExternalService_Networking_Inbound) InboundInterface {
	iface := InboundInterface{
		DataplanePort: inbound.Port,
	}
	if inbound.Address != "" {
		iface.DataplaneIP = inbound.Address
	} else {
		iface.DataplaneIP = n.Address
	}
	if inbound.ServiceAddress != "" {
		iface.WorkloadIP = inbound.ServiceAddress
	} else {
		iface.WorkloadIP = "127.0.0.1"
	}
	if inbound.ServicePort != 0 {
		iface.WorkloadPort = inbound.ServicePort
	} else {
		iface.WorkloadPort = inbound.Port
	}
	return iface
}

// Matches is simply an alias for MatchTags to make source code more aesthetic.
func (d *ExternalService) Matches(selector TagSelector) bool {
	if d != nil {
		return d.MatchTags(selector)
	}
	return false
}

func (d *ExternalService) MatchTags(selector TagSelector) bool {
	for _, inbound := range d.GetNetworking().GetInbound() {
		if inbound.MatchTags(selector) {
			return true
		}
	}
	return false
}

// GetService returns a service represented by this inbound interface.
//
// The purpose of this method is to encapsulate implementation detail
// that service is modeled as a tag rather than a separate field.
func (d *ExternalService_Networking_Inbound) GetService() string {
	if d == nil {
		return ""
	}
	return d.Tags[ServiceTag]
}

// GetProtocol returns a protocol supported by this inbound interface.
//
// The purpose of this method is to encapsulate implementation detail
// that protocol is modeled as a tag rather than a separate field.
func (d *ExternalService_Networking_Inbound) GetProtocol() string {
	if d == nil {
		return ""
	}
	return d.Tags[ProtocolTag]
}

func (d *ExternalService_Networking_Inbound) MatchTags(selector TagSelector) bool {
	return selector.Matches(d.Tags)
}

func (d *ExternalService) Tags() MultiValueTagSet {
	tags := MultiValueTagSet{}
	for _, inbound := range d.GetNetworking().GetInbound() {
		for tag, value := range inbound.Tags {
			_, exists := tags[tag]
			if !exists {
				tags[tag] = map[string]bool{}
			}
			tags[tag][value] = true
		}
	}
	return tags
}

func (d *ExternalService) GetIdentifyingService() string {
	services := d.Tags().Values(ServiceTag)
	if len(services) > 0 {
		return services[0]
	}
	return ServiceUnknown
}
