package mesh

import (
	"fmt"

	mesh_proto "github.com/kumahq/kuma/v3/api/mesh/v1alpha1"
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
			errs = append(errs, fmt.Sprintf("inbound[port=%d] is not ready", inbound.Port))
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
	case !proxyOnline:
		return Offline, errs
	// A proxy that declares neither inbounds nor listeners has nothing to
	// report readiness for and is online once it is connected. That covers an
	// outbound-only proxy and one whose every port is excluded from inbound
	// redirection, both of which used to be reported offline while connected.
	case total == 0:
		return Online, nil
	case ready == 0:
		return Offline, errs
	case ready < total:
		return PartiallyDegraded, errs
	default:
		return Online, nil
	}
}
