package rest

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"github.com/pkg/errors"
	"k8s.io/apiextensions-apiserver/pkg/apiserver/schema/defaulting"
	"k8s.io/kube-openapi/pkg/validation/validate"
	"sigs.k8s.io/yaml"

	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/model/rest/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/resources/registry"
	"github.com/kumahq/kuma/pkg/core/validators"
)

var YAML = &unmarshaler{
	unmarshalFn: func(bytes []byte, i interface{}) error {
		return yaml.Unmarshal(bytes, i)
	},
	marshalFn: yaml.Marshal,
}

var JSON = &unmarshaler{
	unmarshalFn: json.Unmarshal,
	marshalFn:   json.Marshal,
}

type UnmarshalOption struct {
	DisallowStatus bool
}

type UnmarshalOptionFunc func(*UnmarshalOption)

func NewUnmarshalOptions(fs ...UnmarshalOptionFunc) *UnmarshalOption {
	opts := &UnmarshalOption{}
	for _, f := range fs {
		f(opts)
	}
	return opts
}

func DisallowStatus() UnmarshalOptionFunc {
	return func(opts *UnmarshalOption) {
		opts.DisallowStatus = true
	}
}

type unmarshaler struct {
	unmarshalFn func([]byte, interface{}) error
	marshalFn   func(v any) ([]byte, error)
}

type InvalidResourceError struct {
	Reason string
}

func (e *InvalidResourceError) Error() string {
	return e.Reason
}

func (e *InvalidResourceError) Is(target error) bool {
	t, ok := target.(*InvalidResourceError)
	if !ok {
		return false
	}
	return t.Reason == e.Reason || t.Reason == ""
}

func (u *unmarshaler) UnmarshalCore(bytes []byte) (core_model.Resource, error) {
	m := v1alpha1.ResourceMeta{}
	if err := u.unmarshalFn(bytes, &m); err != nil {
		return nil, &InvalidResourceError{Reason: fmt.Sprintf("invalid meta type: %q", err.Error())}
	}
	desc, err := registry.Global().DescriptorFor(core_model.ResourceType(m.Type))
	if err != nil {
		return nil, err
	}
	restResource, err := u.Unmarshal(bytes, desc)
	if err != nil {
		return nil, err
	}
	coreRes, err := To.Core(restResource)
	if err != nil {
		return nil, err
	}
	return coreRes, nil
}

func (u *unmarshaler) Unmarshal(bytes []byte, desc core_model.ResourceTypeDescriptor, fs ...UnmarshalOptionFunc) (Resource, error) {
	resource := desc.NewObject()
	restResource := From.Resource(resource)
	opts := NewUnmarshalOptions(fs...)
	defaultedBytes := bytes
	if desc.Validator != nil && desc.StructuralSchema != nil {
		var err error
		// desc.Schema is set only for new plugin originated policies
		rawObj := map[string]interface{}{}
		// Unfortunately to validate new policies we must first unmarshal into a rawObj
		if err = u.unmarshalFn(bytes, &rawObj); err != nil {
			return nil, &InvalidResourceError{Reason: fmt.Sprintf("invalid %s object: %q", desc.Name, err.Error())}
		}
		if _, found := rawObj["status"]; opts.DisallowStatus && found {
			return nil, &InvalidResourceError{Reason: fmt.Sprintf("invalid %s object: status field cannot be modified", desc.Name)}
		}

		// Apply defaulting
		defaulting.Default(rawObj, desc.StructuralSchema)

		res := desc.Validator.Validate(rawObj)
		if !res.IsValid() {
			return nil, toValidationError(res)
		}
		defaultedBytes, err = u.marshalFn(rawObj)
		if err != nil {
			return nil, err
		}
	}

	if err := u.unmarshalFn(defaultedBytes, restResource); err != nil {
		return nil, &InvalidResourceError{Reason: fmt.Sprintf("invalid %s object: %q", desc.Name, err.Error())}
	}

	if resource.GetMeta() == nil {
		resource.SetMeta(From.Meta(resource))
	}
	if err := core_model.Validate(resource); err != nil {
		return nil, err
	}

	return restResource, nil
}

func (u *unmarshaler) UnmarshalListToCore(b []byte, rs core_model.ResourceList) error {
	rsr := &ResourceListReceiver{
		NewResource: rs.NewItem,
	}
	if err := u.unmarshalFn(b, rsr); err != nil {
		return err
	}
	for _, ri := range rsr.Items {
		r := rs.NewItem()
		if err := r.SetSpec(ri.GetSpec()); err != nil {
			return err
		}
		if r.Descriptor().HasStatus {
			if err := r.SetStatus(ri.GetStatus()); err != nil {
				return err
			}
		}
		r.SetMeta(ri.GetMeta())
		_ = rs.AddItem(r)
	}
	if rsr.Next != nil {
		uri, err := url.ParseRequestURI(*rsr.Next)
		if err != nil {
			return errors.Wrap(err, "invalid next URL from the server")
		}
		offset := uri.Query().Get("offset")
		// we do not preserve here the size of the page, but since it is used in kumactl
		// user will rerun command with the page size of his choice
		if offset != "" {
			rs.GetPagination().SetNextOffset(offset)
		}
	}
	rs.GetPagination().SetTotal(rsr.Total)
	return nil
}

func toValidationError(res *validate.Result) *validators.ValidationError {
	verr := &validators.ValidationError{}
	for _, e := range res.Errors {
		parts := strings.Split(e.Error(), " ")
		if len(parts) > 1 && strings.HasPrefix(parts[0], "spec.") {
			verr.AddViolation(parts[0], strings.Join(parts[1:], " "))
		} else {
			verr.AddViolation("", e.Error())
		}
	}
	return verr
}
