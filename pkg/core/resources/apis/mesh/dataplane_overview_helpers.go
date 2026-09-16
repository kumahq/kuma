package mesh

import (
	"fmt"

	mesh_proto "github.com/kumahq/kuma/v2/api/mesh/v1alpha1"
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

func (t *DataplaneOverviewResource) Status() (Status, []string) {
	proxyOnline := t.Spec.DataplaneInsight.IsOnline()

	var errs []string
	var total int

	for _, inbound := range t.Spec.Dataplane.Networking.Inbound {
		// Ignored inbounds belong to a Service that reaches the Pod only through
		// ignored selector labels (e.g. an Argo Rollouts preview Service), they
		// serve no traffic and must not degrade the proxy.
		if inbound.State == mesh_proto.Dataplane_Networking_Inbound_Ignored {
			continue
		}
		total++
		if (inbound.Health != nil && !inbound.Health.Ready) || inbound.State == mesh_proto.Dataplane_Networking_Inbound_NotReady {
			errs = append(errs, fmt.Sprintf("inbound[port=%d,svc=%s] is not ready", inbound.Port, inbound.Tags[mesh_proto.ServiceTag]))
		}
	}

	allInboundsOffline := len(errs) == total
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
