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
	networking := t.Spec.Dataplane.GetNetworking()

	// Gateway is mutually exclusive with inbounds and zone proxy listeners.
	if networking.GetGateway() != nil {
		if proxyOnline {
			return Online, nil
		}
		return Offline, nil
	}

	var ready, total int
	var errs []string

	for _, inbound := range networking.GetInbound() {
		// Ignored inbounds belong to a Service that reaches the Pod only through
		// ignored selector labels (e.g. an Argo Rollouts preview Service), they
		// serve no traffic and must not degrade the proxy.
		if inbound.State == mesh_proto.Dataplane_Networking_Inbound_Ignored {
			continue
		}
		total++
		if (inbound.Health != nil && !inbound.Health.Ready) || inbound.State == mesh_proto.Dataplane_Networking_Inbound_NotReady {
			errs = append(errs, fmt.Sprintf("inbound[port=%d,svc=%s] is not ready", inbound.Port, inbound.Tags[mesh_proto.ServiceTag]))
		} else {
			ready++
		}
	}

	for _, l := range networking.GetListeners() {
		total++
		if l.State == mesh_proto.Dataplane_Networking_Listener_Ready {
			ready++
		} else {
			errs = append(errs, fmt.Sprintf("listener[port=%d,type=%s] is not ready", l.Port, l.Type))
		}
	}

	switch {
	case !proxyOnline || ready == 0:
		return Offline, errs
	case ready < total:
		return PartiallyDegraded, errs
	default:
		return Online, nil
	}
}
