package rest

import (
	"encoding/json"
	"fmt"
	"net/url"
	"sort"
	"strings"

	"github.com/pkg/errors"
	"k8s.io/apiextensions-apiserver/pkg/apiserver/schema"
	"k8s.io/apiextensions-apiserver/pkg/apiserver/schema/defaulting"
	"k8s.io/apiextensions-apiserver/pkg/apiserver/schema/pruning"
	"k8s.io/kube-openapi/pkg/validation/validate"
	"sigs.k8s.io/yaml"

	core_model "github.com/kumahq/kuma/v3/pkg/core/resources/model"
	"github.com/kumahq/kuma/v3/pkg/core/resources/model/rest/v1alpha1"
	"github.com/kumahq/kuma/v3/pkg/core/resources/registry"
	"github.com/kumahq/kuma/v3/pkg/core/resources/validator"
	"github.com/kumahq/kuma/v3/pkg/core/validators"
)

var YAML = &unmarshaler{
	unmarshalFn: func(bytes []byte, i any) error {
		return yaml.Unmarshal(bytes, i)
	},
	marshalFn: yaml.Marshal,
}

var JSON = &unmarshaler{
	unmarshalFn: json.Unmarshal,
	marshalFn:   json.Marshal,
}

type unmarshaler struct {
	unmarshalFn func([]byte, any) error
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
	return u.unmarshalCore(bytes, false)
}

func (u *unmarshaler) UnmarshalCoreStrict(bytes []byte) (core_model.Resource, error) {
	return u.unmarshalCore(bytes, true)
}

func (u *unmarshaler) unmarshalCore(bytes []byte, strict bool) (core_model.Resource, error) {
	m := v1alpha1.ResourceMeta{}
	if err := u.unmarshalFn(bytes, &m); err != nil {
		return nil, &InvalidResourceError{Reason: fmt.Sprintf("invalid meta type: %q", err.Error())}
	}
	desc, err := registry.Global().DescriptorFor(core_model.ResourceType(m.Type))
	if err != nil {
		return nil, err
	}
	restResource, err := u.unmarshal(bytes, desc, strict)
	if err != nil {
		return nil, err
	}
	coreRes, err := To.Core(restResource)
	if err != nil {
		return nil, err
	}
	return coreRes, nil
}

func (u *unmarshaler) Unmarshal(bytes []byte, desc core_model.ResourceTypeDescriptor) (Resource, error) {
	return u.unmarshal(bytes, desc, false)
}

func (u *unmarshaler) UnmarshalStrict(bytes []byte, desc core_model.ResourceTypeDescriptor) (Resource, error) {
	return u.unmarshal(bytes, desc, true)
}

func (u *unmarshaler) unmarshal(bytes []byte, desc core_model.ResourceTypeDescriptor, strict bool) (Resource, error) {
	resource := desc.NewObject()
	restResource := From.Resource(resource)
	defaultedBytes := bytes
	if desc.Validator != nil && desc.StructuralSchema != nil {
		var err error
		// desc.Schema is set only for new plugin originated policies
		rawObj := map[string]any{}
		// Unfortunately to validate new policies we must first unmarshal into a rawObj
		if err = u.unmarshalFn(bytes, &rawObj); err != nil {
			return nil, &InvalidResourceError{Reason: fmt.Sprintf("invalid %s object: %q", desc.Name, err.Error())}
		}

		if strict {
			if err := rejectUnknownFields(rawObj, desc.StructuralSchema); err != nil {
				return nil, err
			}
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
	if err := validator.Validate(resource); err != nil {
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

func rejectUnknownFields(rawObj map[string]any, structuralSchema *schema.Structural) error {
	unknownFields := pruning.PruneWithOptions(rawObj, structuralSchema, true, schema.UnknownFieldPathOptions{TrackUnknownFieldPaths: true})
	for _, k8sField := range []string{"apiVersion", "kind", "metadata"} {
		if _, ok := rawObj[k8sField]; !ok {
			continue
		}
		if _, known := structuralSchema.Properties[k8sField]; known {
			continue
		}
		unknownFields = append(unknownFields, k8sField)
	}
	if len(unknownFields) == 0 {
		return nil
	}
	sort.Strings(unknownFields)
	verr := &validators.ValidationError{}
	for _, field := range unknownFields {
		verr.AddViolation(field, "unknown field")
	}
	return verr
}

func toValidationError(res *validate.Result) *validators.ValidationError {
	verr := &validators.ValidationError{}
	// res.Errors order is not deterministic: the underlying schema validator
	// walks JSONSchemaProps.Properties, a Go map, so errors for sibling
	// fields can come back in a different order on every call. Sort by
	// message so the resulting violations are stable across runs.
	errs := make([]string, len(res.Errors))
	for i, e := range res.Errors {
		errs[i] = e.Error()
	}
	sort.Strings(errs)
	for _, e := range errs {
		parts := strings.Split(e, " ")
		if len(parts) > 1 && strings.HasPrefix(parts[0], "spec.") {
			verr.AddViolation(parts[0], strings.Join(parts[1:], " "))
		} else {
			verr.AddViolation("", e)
		}
	}
	return verr
}
