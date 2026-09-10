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

	var ready int
	var errs []string
	total := len(networking.GetInbound()) + len(networking.GetListeners())

	for _, inbound := range networking.GetInbound() {
		if (inbound.Health != nil && !inbound.Health.Ready) || inbound.State == mesh_proto.Dataplane_Networking_Inbound_NotReady {
			errs = append(errs, fmt.Sprintf("inbound[port=%d] is not ready", inbound.Port))
		} else {
			ready++
		}
	}

	for _, l := range networking.GetListeners() {
		if l.State == mesh_proto.Dataplane_Networking_Listener_Ready {
			ready++
		} else {
			errs = append(errs, fmt.Sprintf("listener[port=%d,type=%s] is not ready", l.Port, l.Type))
		}
	}

	switch {
	case !proxyOnline:
		return Offline, errs
	// A proxy that declares neither inbounds nor listeners, such as one that
	// only fronts traffic on ports excluded from inbound redirection, has
	// nothing to report readiness for and is online once it is connected.
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
