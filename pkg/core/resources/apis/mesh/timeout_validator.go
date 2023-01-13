package mesh

import (
	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/validators"
)

func (t *TimeoutResource) Validate() error {
	var err validators.ValidationError

	err.Add(t.validateSources())
	err.Add(t.validateDestinations())
	err.Add(t.validateConf())

	return err.OrNil()
}

func (t *TimeoutResource) validateSources() validators.ValidationError {
	return ValidateSelectors(
		validators.RootedAt("sources"),
		t.Spec.Sources,
		ValidateSelectorsOpts{
			ValidateTagsOpts: ValidateTagsOpts{
				RequireAtLeastOneTag: true,
				RequireService:       true,
			},
			RequireAtLeastOneSelector: true,
		},
	)
}

func (t *TimeoutResource) validateDestinations() validators.ValidationError {
	return ValidateSelectors(
		validators.RootedAt("destinations"),
		t.Spec.Destinations,
		OnlyServiceTagAllowed,
	)
}

func (t *TimeoutResource) validateConf() validators.ValidationError {
	var err validators.ValidationError
	path := validators.RootedAt("conf")
	conf := t.Spec.GetConf()

	if conf == nil {
		err.AddViolationAt(path, HasToBeDefinedViolation)
		return err
	}

	err.Add(t.validateConfTcp(path.Field("tcp"), conf.GetTcp()))
	err.Add(t.validateConfHttp(path.Field("http"), conf.GetHttp()))
	err.Add(t.validateConfGrpc(path.Field("grpc"), conf.GetGrpc()))

	return err
}

func (t *TimeoutResource) validateConfTcp(path validators.PathBuilder, conf *mesh_proto.Timeout_Conf_Tcp) validators.ValidationError {
	var err validators.ValidationError
	if conf == nil {
		return err
	}
	if conf.IdleTimeout == nil {
		err.AddViolationAt(path, "at least one timeout in section has to be defined")
		return err
	}
	return validateDuration_GreaterThan0(path.Field("idleTimeout"), conf.IdleTimeout)
}

func (t *TimeoutResource) validateConfHttp(path validators.PathBuilder, conf *mesh_proto.Timeout_Conf_Http) validators.ValidationError {
	var err validators.ValidationError
	if conf == nil {
		return err
	}
	if conf.RequestTimeout == nil && conf.IdleTimeout == nil {
		err.AddViolationAt(path, "at least one timeout in section has to be defined")
		return err
	}
	err.Add(validateDuration_GreaterThan0OrNil(path.Field("requestTimeout"), conf.RequestTimeout))
	err.Add(validateDuration_GreaterThan0OrNil(path.Field("idleTimeout"), conf.IdleTimeout))
	return err
}

func (t *TimeoutResource) validateConfGrpc(path validators.PathBuilder, conf *mesh_proto.Timeout_Conf_Grpc) validators.ValidationError {
	var err validators.ValidationError
	if conf == nil {
		return err
	}
	if conf.StreamIdleTimeout == nil && conf.MaxStreamDuration == nil {
		err.AddViolationAt(path, "at least one timeout in section has to be defined")
		return err
	}
	err.Add(validateDuration_GreaterThan0OrNil(path.Field("streamIdleTimeout"), conf.StreamIdleTimeout))
	err.Add(validateDuration_GreaterThan0OrNil(path.Field("maxStreamDuration"), conf.MaxStreamDuration))
	return err
}
