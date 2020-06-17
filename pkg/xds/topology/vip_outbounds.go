package topology

import (
	"strings"

	"github.com/pkg/errors"
	"go.uber.org/multierr"

	mesh_proto "github.com/Kong/kuma/api/mesh/v1alpha1"
	mesh_core "github.com/Kong/kuma/pkg/core/resources/apis/mesh"
	"github.com/Kong/kuma/pkg/dns-server/resolver"
)

const VIPListenPort = uint32(80)

func PatchDataplaneWithVIPOutbounds(dataplane *mesh_core.DataplaneResource,
	dataplanes *mesh_core.DataplaneResourceList, resolver resolver.DNSResolver) (errs error) {
	serviceVIPMap := map[string]string{}
	for _, dp := range dataplanes.Items {
		if dp.Meta.GetName() == dataplane.Meta.GetName() {
			continue
		}

		for _, inbound := range dp.Spec.Networking.Inbound {
			inService := inbound.GetTags()[mesh_proto.ServiceTag]

			vip, err := resolver.ForwardLookup(inService)
			if err != nil {
				// try to get the first part of the FQDN service and look it up
				split := strings.Split(inService, ".")
				vip, err = resolver.ForwardLookup(split[0])
				if err != nil {
					errs = multierr.Append(errs, errors.Wrapf(err, "unable to resolve %s", inService))
				}
			}
			serviceVIPMap[inService] = vip
		}
	}

	for service, vip := range serviceVIPMap {
		dataplane.Spec.Networking.Outbound = append(dataplane.Spec.Networking.Outbound,
			&mesh_proto.Dataplane_Networking_Outbound{
				Address: vip,
				Port:    VIPListenPort,
				Service: service,
			})
	}

	return
}
