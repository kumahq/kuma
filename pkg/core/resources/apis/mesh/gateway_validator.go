package mesh

import (
	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/validators"
)

// Validate checks MeshGatewayResource semantic constraints.
func (g *MeshGatewayResource) Validate() error {
	var err validators.ValidationError

	onlyOneSelector := g.Spec.IsCrossMesh()

	err.Add(ValidateSelectors(
		validators.RootedAt("selectors"),
		g.Spec.GetSelectors(),
		ValidateSelectorsOpts{
			RequireAtLeastOneSelector: true,
			RequireAtMostOneSelector:  onlyOneSelector,
			ValidateTagsOpts: ValidateTagsOpts{
				RequireAtLeastOneTag: true,
				RequireService:       true,
			},
		},
	))

	// The top-level selector is used to bind the gateway to a set
	// of dataplanes, so the service tag must not be used here.
	err.Add(ValidateSelector(
		validators.RootedAt("tags"),
		g.Spec.GetTags(),
		ValidateTagsOpts{
			ExtraTagKeyValidators: []TagKeyValidatorFunc{
				SelectorKeyNotInSet(
					mesh_proto.ExternalServiceTag,
					mesh_proto.ProtocolTag,
					mesh_proto.ServiceTag,
					mesh_proto.ZoneTag,
				),
			},
		},
	))

	err.Add(validateMeshGatewayConf(
		validators.RootedAt("conf"),
		g.Spec.GetConf(),
	))

	return err.OrNil()
}

type resourceLimits struct {
	connectionLimits map[uint32]struct{}
	listeners        []int
}

func validateListenerCollapsibility(path validators.PathBuilder, listeners []*mesh_proto.MeshGateway_Listener) validators.ValidationError {
	protocolsForPort := map[uint32]map[string][]int{}
	hostnamesForPort := map[uint32]map[string][]int{}
	limitedListenersForPort := map[uint32]resourceLimits{}

	for i, listener := range listeners {
		protocols, ok := protocolsForPort[listener.GetPort()]
		if !ok {
			protocols = map[string][]int{}
		}

		hostnames, ok := hostnamesForPort[listener.GetPort()]
		if !ok {
			hostnames = map[string][]int{}
		}

		limitedListeners, ok := limitedListenersForPort[listener.GetPort()]
		if !ok {
			limitedListeners = resourceLimits{
				connectionLimits: map[uint32]struct{}{},
			}
		}

		protocols[listener.GetProtocol().String()] = append(protocols[listener.GetProtocol().String()], i)

		// An empty hostname is the same as "*", i.e. matches all hosts.
		hostname := listener.GetNonEmptyHostname()

		hostnames[hostname] = append(hostnames[hostname], i)

		if l := listener.GetResources().GetConnectionLimit(); l != 0 {
			limitedListeners.listeners = append(limitedListeners.listeners, i)
			limitedListeners.connectionLimits[l] = struct{}{}
		}

		hostnamesForPort[listener.GetPort()] = hostnames
		protocolsForPort[listener.GetPort()] = protocols
		limitedListenersForPort[listener.GetPort()] = limitedListeners
	}

	err := validators.ValidationError{}

	for _, protocolIndexes := range protocolsForPort {
		if len(protocolIndexes) <= 1 {
			continue
		}

		for _, indexes := range protocolIndexes {
			for _, index := range indexes {
				err.AddViolationAt(path.Index(index), "protocol conflicts with other listeners on this port")
			}
		}
	}

	for _, hostnameIndexes := range hostnamesForPort {
		for _, indexes := range hostnameIndexes {
			if len(indexes) <= 1 {
				continue
			}

			for _, index := range indexes {
				err.AddViolationAt(path.Index(index), "multiple listeners for hostname on this port")
			}
		}
	}

	for _, listeners := range limitedListenersForPort {
		if len(listeners.connectionLimits) <= 1 {
			continue
		}
		for _, index := range listeners.listeners {
			err.AddViolationAt(path.Index(index).Field("resources").Field("connectionLimit"), "conflicting values for this port")
		}
	}

	return err
}

func validateMeshGatewayConf(path validators.PathBuilder, conf *mesh_proto.MeshGateway_Conf) validators.ValidationError {
	err := validators.ValidationError{}

	if conf == nil {
		err.AddViolationAt(path, "cannot be empty")
		return err
	}

	path = path.Field("listeners")

	if len(conf.GetListeners()) == 0 {
		err.AddViolationAt(path, "cannot be empty")
	}

	for i, l := range conf.GetListeners() {
		// Hostname is optional, since it might be given on the route(s).
		if l.GetHostname() != "" {
			err.Add(ValidateHostname(path.Index(i).Field("hostname"), l.GetHostname()))
		}

		// Port is required, and must not be 0.
		err.Add(ValidatePort(path.Index(i).Field("port"), l.GetPort()))

		// For now, only support HTTP and HTTPS.
		switch l.GetProtocol() {
		case mesh_proto.MeshGateway_Listener_NONE:
			err.AddViolationAt(path.Index(i).Field("protocol"), "cannot be empty")
		case mesh_proto.MeshGateway_Listener_HTTPS:
			switch {
			case l.GetCrossMesh():
				err.AddViolationAt(path.Index(i).Field("protocol"), "protocol is not supported with crossMesh")
			case l.GetTls() == nil:
				err.AddViolationAt(path.Index(i).Field("tls"), "cannot be empty")
			case l.GetTls().GetMode() == mesh_proto.MeshGateway_TLS_PASSTHROUGH:
				err.AddViolationAt(
					path.Index(i).Field("tls").Field("mode"),
					"mode is not supported on HTTPS listeners")
			}
		}

		if tls := l.GetTls(); tls != nil && !l.GetCrossMesh() {
			switch tls.GetMode() {
			case mesh_proto.MeshGateway_TLS_NONE:
				err.AddViolationAt(
					path.Index(i).Field("tls").Field("mode"),
					"cannot be empty")
			case mesh_proto.MeshGateway_TLS_TERMINATE:
				switch len(tls.GetCertificates()) {
				case 0:
					err.AddViolationAt(
						path.Index(i).Field("tls").Field("certificates"),
						"cannot be empty in TLS termination mode")
				case 1, 2:
					// Can have RSA and/or ECDSA certificates.
				default:
					err.AddViolationAt(
						path.Index(i).Field("tls").Field("certificates"),
						"cannot have more than 2 certificates")
				}
			}
		}
		if tls := l.GetTls(); tls != nil && l.GetCrossMesh() {
			if tls.GetMode() != mesh_proto.MeshGateway_TLS_NONE ||
				len(tls.GetCertificates()) > 0 {
				err.AddViolationAt(
					path.Index(i).Field("tls"),
					"must be empty with crossMesh")
			}
		}

		// Listener tags are optional, but if given, must not contain
		// various tags that are well-known properties of Dataplanes.
		err.Add(ValidateSelector(
			path.Index(i).Field("tags"),
			l.GetTags(),
			ValidateTagsOpts{
				ExtraTagKeyValidators: []TagKeyValidatorFunc{
					SelectorKeyNotInSet(
						mesh_proto.ExternalServiceTag,
						mesh_proto.ProtocolTag,
						mesh_proto.ServiceTag,
						mesh_proto.ZoneTag,
					),
				},
			}))
	}

	err.Add(validateListenerCollapsibility(path, conf.GetListeners()))

	return err
}
