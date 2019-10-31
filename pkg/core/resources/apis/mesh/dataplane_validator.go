package mesh

import (
	mesh_proto "github.com/Kong/kuma/api/mesh/v1alpha1"
	"github.com/Kong/kuma/pkg/core/validators"
	"github.com/pkg/errors"
	"go.uber.org/multierr"
)

func (d *DataplaneResource) Validate() error {
	var errs error
	for i, inbound := range d.Spec.GetNetworking().GetInbound() {
		if err := validateInbound(inbound); err != nil {
			errs = multierr.Combine(errs, errors.Wrapf(err, "Inbound[%d]", i))
		}
	}
	for i, outbound := range d.Spec.GetNetworking().GetOutbound() {
		if err := validateOutbound(outbound); err != nil {
			errs = multierr.Combine(errs, errors.Wrapf(err, "Outbound[%d]", i))
		}
	}
	return validators.NewValidationError(errs)
}

func validateInbound(inbound *mesh_proto.Dataplane_Networking_Inbound) error {
	var errs error
	if _, err := mesh_proto.ParseInboundInterface(inbound.Interface); err != nil {
		errs = multierr.Combine(errs, errors.New("Interface: invalid format: expected format is <DATAPLANE_IP>:<DATAPLANE_PORT>:<WORKLOAD_PORT> ex. 192.168.0.100:9090:8080"))
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
