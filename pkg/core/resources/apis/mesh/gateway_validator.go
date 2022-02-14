package mesh

import (
	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/validators"
)

// Validate checks MeshGatewayResource semantic constraints.
func (g *MeshGatewayResource) Validate() error {
	var err validators.ValidationError

	err.Add(ValidateSelectors(
		validators.RootedAt("selectors"),
		g.Spec.GetSelectors(),
		ValidateSelectorsOpts{
			RequireAtLeastOneSelector: true,
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
		case mesh_proto.MeshGateway_Listener_UDP,
			mesh_proto.MeshGateway_Listener_TCP,
			mesh_proto.MeshGateway_Listener_TLS:
			err.AddViolationAt(path.Index(i).Field("protocol"), "protocol type is not supported")
		case mesh_proto.MeshGateway_Listener_HTTPS:
			if l.GetTls() == nil {
				err.AddViolationAt(path.Index(i).Field("tls"), "cannot be empty")
			}
		}

		if tls := l.GetTls(); tls != nil {
			switch tls.GetMode() {
			case mesh_proto.MeshGateway_TLS_NONE:
				err.AddViolationAt(
					path.Index(i).Field("tls").Field("mode"),
					"cannot be empty")
			case mesh_proto.MeshGateway_TLS_PASSTHROUGH:
				if len(tls.GetCertificates()) > 0 {
					err.AddViolationAt(
						path.Index(i).Field("tls").Field("certificates"),
						"must be empty in TLS passthrough mode")
				}
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

	return err
}
