package mesh

import (
	"fmt"
	"regexp"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/validators"
)

func (t *VirtualOutboundResource) Validate() error {
	var err validators.ValidationError
	err.Add(t.validateSelectors())
	err.Add(t.ValidateConf())
	return err.OrNil()
}

func (t *VirtualOutboundResource) validateSelectors() validators.ValidationError {
	return ValidateSelectors(validators.RootedAt("selectors"), t.Spec.GetSelectors(), ValidateSelectorsOpts{
		RequireAtLeastOneSelector: true,
		ValidateTagsOpts: ValidateTagsOpts{
			RequireAtLeastOneTag: true,
		},
	})
}

func (t *VirtualOutboundResource) ValidateConf() (err validators.ValidationError) {
	root := validators.RootedAt("conf")
	// host, port, parameters
	conf := t.Spec.GetConf()

	if conf == nil {
		err.AddViolationAt(root, HasToBeDefinedViolation)
		return
	}
	err.Add(t.validateParameters(root.Field("parameters")))
	err.Add(t.validateHost(root.Field("host")))
	err.Add(t.validatePort(root.Field("port")))
	return
}

var parameterKeyName = regexp.MustCompile(`^[a-zA-Z0-9]+$`)

func (t *VirtualOutboundResource) validateParameters(path validators.PathBuilder) (err validators.ValidationError) {
	hasService := false
	keys := map[string]bool{}
	for k, v := range t.Spec.Conf.Parameters {
		p := path.Index(k)
		if keys[v.Name] {
			err.AddViolationAt(p.Field("name"), `name is already used`)
		}
		keys[v.Name] = true
		if v.TagKey != "" {
			hasService = hasService || v.TagKey == mesh_proto.ServiceTag
			if !tagNameCharacterSet.MatchString(v.TagKey) {
				err.AddViolationAt(p.Field("tagKey"), `must be a valid tag name`)
			}
		} else if !tagNameCharacterSet.MatchString(v.Name) {
			err.AddViolationAt(p.Field("name"), `must be a valid tag name if you are not setting tagKey`)
		}
		if !parameterKeyName.MatchString(v.Name) {
			err.AddViolationAt(p.Field("name"), `must consist of alphanumeric characters to be used as a gotemplate key`)
		}
	}
	if !hasService {
		err.AddViolationAt(path, fmt.Sprintf(`must contain a parameter with %s as a tagKey`, mesh_proto.ServiceTag))
	}
	return
}

func (t *VirtualOutboundResource) validateHost(path validators.PathBuilder) (err validators.ValidationError) {
	h := t.Spec.Conf.Host
	if h == "" {
		err.AddViolationAt(path, HasToBeDefinedViolation)
		return
	}
	fakeTags := map[string]string{}
	for _, v := range t.Spec.Conf.Parameters {
		fakeTags[tagKeyOrName(v)] = "dummy"
	}
	_, lerr := t.EvalHost(fakeTags)
	if lerr != nil {
		err.AddViolationAt(path, fmt.Sprintf("template pre evaluation failed with error='%s'", lerr.Error()))
	}
	return
}

func (t *VirtualOutboundResource) validatePort(path validators.PathBuilder) (err validators.ValidationError) {
	h := t.Spec.Conf.Port
	if h == "" {
		err.AddViolationAt(path, HasToBeDefinedViolation)
		return
	}
	fakeTags := map[string]string{}
	for _, v := range t.Spec.Conf.Parameters {
		fakeTags[tagKeyOrName(v)] = "1"
	}
	_, lerr := t.EvalPort(fakeTags)
	if lerr != nil {
		err.AddViolationAt(path, fmt.Sprintf("template pre evaluation failed with error='%s'", lerr.Error()))
	}
	return
}
