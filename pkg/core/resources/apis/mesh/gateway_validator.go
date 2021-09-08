package mesh

import (
	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/validators"
)

// Validate checks GatewayResource semantic constraints.
func (g *GatewayResource) Validate() error {
	var err validators.ValidationError

	err.Add(ValidateSelectors(
		validators.RootedAt("selectors"),
		g.Spec.GetSelectors(),
		ValidateSelectorsOpts{
			RequireAtLeastOneSelector: true,
			ValidateSelectorOpts: ValidateSelectorOpts{
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
		ValidateSelectorOpts{
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

	err.Add(validateGatewayConf(
		validators.RootedAt("conf"),
		g.Spec.GetConf(),
	))

	return err.OrNil()
}

func validateGatewayConf(path validators.PathBuilder, conf *mesh_proto.Gateway_Conf) validators.ValidationError {
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
		case mesh_proto.Gateway_Listener_NONE:
			err.AddViolationAt(path.Index(i).Field("protocol"), "cannot be empty")
		case mesh_proto.Gateway_Listener_UDP,
			mesh_proto.Gateway_Listener_TCP,
			mesh_proto.Gateway_Listener_TLS:
			err.AddViolationAt(path.Index(i).Field("protocol"), "protocol type is not supported")
		}

		if tls := l.GetTls(); tls != nil {
			switch tls.GetMode() {
			case mesh_proto.Gateway_TLS_NONE:
				err.AddViolationAt(
					path.Index(i).Field("tls").Field("mode"),
					"cannot be empty")
			case mesh_proto.Gateway_TLS_PASSTHROUGH:
				if tls.GetCertificate() != nil {
					err.AddViolationAt(
						path.Index(i).Field("tls").Field("certificate"),
						"must be empty in TLS passthrough mode")
				}
			case mesh_proto.Gateway_TLS_TERMINATE:
				if tls.GetCertificate() == nil {
					err.AddViolationAt(
						path.Index(i).Field("tls").Field("certificate"),
						"cannot be empty in TLS termination mode")
				}
			}
		}

		err.Add(ValidateSelector(
			path.Index(i).Field("tags"),
			l.GetTags(),
			ValidateSelectorOpts{
				RequireAtLeastOneTag: true,
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
