package mesh

import (
	"reflect"

	mesh_proto "github.com/Kong/kuma/api/mesh/v1alpha1"
	"github.com/Kong/kuma/pkg/core/validators"
)

func (r *HealthCheckResource) HasActiveChecks() bool {
	activeChecks := r.Spec.Conf.GetActiveChecks()
	return activeChecks != nil && !reflect.DeepEqual(*activeChecks, mesh_proto.HealthCheck_Conf_Active{})
}

func (r *HealthCheckResource) HasPassiveChecks() bool {
	passiveChecks := r.Spec.Conf.GetPassiveChecks()
	return passiveChecks != nil && !reflect.DeepEqual(*passiveChecks, mesh_proto.HealthCheck_Conf_Passive{})
}

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
		ValidateSelectorOpts: ValidateSelectorOpts{
			RequireAtLeastOneTag: true,
			RequireService:       true,
		},
	})
}

func (d *HealthCheckResource) validateDestinations() (err validators.ValidationError) {
	return ValidateSelectors(validators.RootedAt("destinations"), d.Spec.Destinations, OnlyServiceTagAllowed)
}

func (d *HealthCheckResource) validateConf() (err validators.ValidationError) {
	root := validators.RootedAt("conf")
	if !d.HasActiveChecks() && !d.HasPassiveChecks() {
		err.AddViolationAt(root, "must have either active or passive checks configured")
	}
	if d.HasActiveChecks() {
		path := root.Field("activeChecks")
		activeChecks := d.Spec.Conf.GetActiveChecks()
		err.Add(ValidateDuration(path.Field("interval"), activeChecks.Interval))
		err.Add(ValidateDuration(path.Field("timeout"), activeChecks.Timeout))
		err.Add(ValidateThreshold(path.Field("unhealthyThreshold"), activeChecks.UnhealthyThreshold))
		err.Add(ValidateThreshold(path.Field("healthyThreshold"), activeChecks.HealthyThreshold))
	}
	if d.HasPassiveChecks() {
		path := root.Field("passiveChecks")
		passiveChecks := d.Spec.Conf.GetPassiveChecks()
		err.Add(ValidateThreshold(path.Field("unhealthyThreshold"), passiveChecks.UnhealthyThreshold))
		err.Add(ValidateDuration(path.Field("penaltyInterval"), passiveChecks.PenaltyInterval))
	}
	return
}
