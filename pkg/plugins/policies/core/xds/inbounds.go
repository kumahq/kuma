package xds

import (
	mesh_proto "github.com/kumahq/kuma/v3/api/mesh/v1alpha1"
	core_mesh "github.com/kumahq/kuma/v3/pkg/core/resources/apis/mesh"
	core_rules "github.com/kumahq/kuma/v3/pkg/plugins/policies/core/rules"
	rules_inbound "github.com/kumahq/kuma/v3/pkg/plugins/policies/core/rules/inbound"
)

// InboundMatch is a single dataplane inbound together with the inbound rules
// matched for it and their merged catch-all conf.
type InboundMatch[T any] struct {
	Inbound   *mesh_proto.Dataplane_Networking_Inbound
	Interface mesh_proto.InboundInterface
	Listener  core_rules.InboundListener
	Rules     []*rules_inbound.Rule
	Conf      T
}

// ForEachInbound walks every inbound declared on the dataplane, resolving the
// inbound rules matched for it and the merged catch-all conf of type T.
// Callers are responsible for looking up any xDS resource (listener, cluster)
// keyed by InboundMatch.Listener.
func ForEachInbound[T any](dataplane *core_mesh.DataplaneResource, fromRules core_rules.FromRules, fn func(InboundMatch[T]) error) error {
	for _, inbound := range dataplane.Spec.GetNetworking().GetInbound() {
		iface := dataplane.Spec.Networking.ToInboundInterface(inbound)

		listenerKey := core_rules.InboundListener{
			Address: iface.DataplaneIP,
			Port:    iface.DataplanePort,
		}

		rules := fromRules.InboundRules[listenerKey]
		conf := rules_inbound.MatchesAllIncomingTraffic[T](rules)

		if err := fn(InboundMatch[T]{
			Inbound:   inbound,
			Interface: iface,
			Listener:  listenerKey,
			Rules:     rules,
			Conf:      conf,
		}); err != nil {
			return err
		}
	}

	return nil
}
