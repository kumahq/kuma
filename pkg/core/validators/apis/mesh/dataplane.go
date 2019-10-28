package mesh

import (
	mesh_proto "github.com/Kong/kuma/api/mesh/v1alpha1"
	"github.com/Kong/kuma/pkg/core/resources/apis/mesh"
	"github.com/Kong/kuma/pkg/core/validators"
	"github.com/pkg/errors"
	"go.uber.org/multierr"
)

func ValidateDataplane(dataplane *mesh.DataplaneResource) error {
	var errs error
	for i, inbound := range dataplane.Spec.GetNetworking().GetInbound() {
		if err := validateInbound(inbound); err != nil {
			errs = multierr.Combine(errs, errors.Wrapf(err, "Inbound[%d]", i))
		}
	}
	for i, outbound := range dataplane.Spec.GetNetworking().GetOutbound() {
		if err := validateOutbound(outbound); err != nil {
			errs = multierr.Combine(errs, errors.Wrapf(err, "Outbound[%d]", i))
		}
	}
	return validators.NewValidationError(errs)
}

func validateInbound(inbound *mesh_proto.Dataplane_Networking_Inbound) error {
	var errs error
	if _, err := mesh_proto.ParseInboundInterface(inbound.Interface); err != nil {
		errs = multierr.Combine(errs, errors.Wrap(err, "Interface"))
	}
	tag := inbound.Tags[mesh_proto.ServiceTag]
	if tag == "" {
		errs = multierr.Combine(errs, errors.New(`"service" tag has to exist and be non empty`))
	}
	return errs
}

func validateOutbound(outbound *mesh_proto.Dataplane_Networking_Outbound) error {
	var errs error
	if _, err := mesh_proto.ParseOutboundInterface(outbound.Interface); err != nil {
		errs = multierr.Combine(errs, errors.Wrap(err, "Interface"))
	}
	if outbound.Service == "" {
		errs = multierr.Combine(errs, errors.New("Service cannot be empty"))
	}
	return errs
}
