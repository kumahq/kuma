package mesh

import (
	"fmt"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
)

type Status string

func (s Status) String() string {
	return string(s)
}

const (
	Online            = Status("Online")
	Offline           = Status("Offline")
	PartiallyDegraded = Status("Partially degraded")
)

func (t *DataplaneOverviewResource) GetStatus() (Status, []string) {
	proxyOnline := t.Spec.DataplaneInsight.IsOnline()

	var errs []string

	for _, inbound := range t.Spec.Dataplane.Networking.Inbound {
		if inbound.Health != nil && !inbound.Health.Ready {
			errs = append(errs, fmt.Sprintf("inbound[port=%d,svc=%s] is not ready", inbound.Port, inbound.Tags[mesh_proto.ServiceTag]))
		}
	}

	allInboundsOffline := len(errs) == len(t.Spec.Dataplane.Networking.Inbound)
	allInboundsOnline := len(errs) == 0

	if t.Spec.Dataplane.GetNetworking().GetGateway() != nil {
		allInboundsOffline = false
		allInboundsOnline = true
	}

	if !proxyOnline || allInboundsOffline {
		return Offline, errs
	}
	if !allInboundsOnline {
		return PartiallyDegraded, errs
	}
	return Online, nil
}
