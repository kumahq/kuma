package mesh

import (
	"reflect"

	mesh_proto "github.com/Kong/kuma/api/mesh/v1alpha1"
	"github.com/Kong/kuma/pkg/core/validators"
)

func (d *HealthCheckResource) Validate() error {
	var err validators.ValidationError
	err.Add(d.validateSources())
	err.Add(d.validateDestinations())
	err.Add(d.validateConf())
	return err.OrNil()
}

func (d *HealthCheckResource) validateSources() validators.ValidationError {
	return ValidateSelectors(validators.RootedAt("sources"), d.Spec.Sources, ValidateSelectorsOpts{
		RequireAtLeastOneSelector: true,
	})
}

func (d *HealthCheckResource) validateDestinations() (err validators.ValidationError) {
	return ValidateSelectors(validators.RootedAt("destinations"), d.Spec.Destinations, OnlyServiceTagAllowed)
}

func (d *HealthCheckResource) validateConf() (err validators.ValidationError) {
	root := validators.RootedAt("conf")
	activeChecks := d.Spec.Conf.GetActiveChecks()
	passiveChecks := d.Spec.Conf.GetPassiveChecks()
	hasActiveChecks := activeChecks != nil && !reflect.DeepEqual(*activeChecks, mesh_proto.HealthCheck_Conf_Active{})
	hasPassiveChecks := passiveChecks != nil && !reflect.DeepEqual(*passiveChecks, mesh_proto.HealthCheck_Conf_Passive{})
	if !hasActiveChecks && !hasPassiveChecks {
		err.AddViolationAt(root, "must have either active or passive checks configured")
	}
	if hasActiveChecks {
		path := root.Field("activeChecks")
		err.Add(ValidateDuration(path.Field("interval"), activeChecks.GetInterval()))
		err.Add(ValidateDuration(path.Field("timeout"), activeChecks.GetTimeout()))
		err.Add(ValidateThreshold(path.Field("unhealthyThreshold"), activeChecks.UnhealthyThreshold))
		err.Add(ValidateThreshold(path.Field("healthyThreshold"), activeChecks.HealthyThreshold))
	}
	if hasPassiveChecks {
		path := root.Field("passiveChecks")
		err.Add(ValidateThreshold(path.Field("unhealthyThreshold"), passiveChecks.UnhealthyThreshold))
		err.Add(ValidateDuration(path.Field("penaltyInterval"), passiveChecks.GetPenaltyInterval()))
	}
	return
}
