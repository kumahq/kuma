package mesh

import (
	"fmt"
	"regexp"

	"github.com/kumahq/kuma/pkg/core/validators"
)

func (r *VirtualOutboundResource) Validate() error {
	var err validators.ValidationError
	err.Add(r.validateSelectors())
	err.Add(r.ValidateConf())
	return err.OrNil()
}

func (r *VirtualOutboundResource) validateSelectors() validators.ValidationError {
	return ValidateSelectors(validators.RootedAt("selectors"), r.Spec.GetSelectors(), ValidateSelectorsOpts{
		RequireAtLeastOneSelector: true,
		ValidateSelectorOpts: ValidateSelectorOpts{
			RequireAtLeastOneTag: true,
		},
	})
}

func (r *VirtualOutboundResource) ValidateConf() (err validators.ValidationError) {
	root := validators.RootedAt("conf")
	// host, port, parameters
	conf := r.Spec.GetConf()

	if conf == nil {
		err.AddViolationAt(root, HasToBeDefinedViolation)
		return
	}
	err.Add(r.validateParameters(root.Field("parameters")))
	err.Add(r.validateHost(root.Field("host")))
	err.Add(r.validatePort(root.Field("port")))
	return
}

var parameterKeyName = regexp.MustCompile(`^[a-zA-Z0-9]+$`)

func (r *VirtualOutboundResource) validateParameters(path validators.PathBuilder) (err validators.ValidationError) {
	for k, v := range r.Spec.Conf.Parameters {
		if !tagNameCharacterSet.MatchString(v) {
			err.AddViolationAt(path.Key(k), `value of parameters must be a valid tag name`)
		}
		if !parameterKeyName.MatchString(k) {
			err.AddViolationAt(path.Key(k), `key of parameters must consist of alphanumeric characters`)
		}
	}
	return
}

func (r *VirtualOutboundResource) validateHost(path validators.PathBuilder) (err validators.ValidationError) {
	h := r.Spec.Conf.Host
	if h == "" {
		err.AddViolationAt(path, HasToBeDefinedViolation)
		return
	}
	fakeTags := map[string]string{}
	for _, v := range r.Spec.Conf.Parameters {
		fakeTags[v] = "dummy"
	}
	_, lerr := r.EvalHost(fakeTags)
	if lerr != nil {
		err.AddViolationAt(path, fmt.Sprintf("template pre evaluation failed with error='%s'", lerr.Error()))
	}
	return
}

func (r *VirtualOutboundResource) validatePort(path validators.PathBuilder) (err validators.ValidationError) {
	h := r.Spec.Conf.Port
	if h == "" {
		err.AddViolationAt(path, HasToBeDefinedViolation)
		return
	}
	fakeTags := map[string]string{}
	for _, v := range r.Spec.Conf.Parameters {
		fakeTags[v] = "1"
	}
	_, lerr := r.EvalPort(fakeTags)
	if lerr != nil {
		err.AddViolationAt(path, fmt.Sprintf("template pre evaluation failed with error='%s'", lerr.Error()))
	}
	return
}
