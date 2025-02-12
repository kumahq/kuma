package rest

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/google/go-cmp/cmp"
	"net/url"
	"reflect"
	"strings"

	"github.com/pkg/errors"
	"k8s.io/apiextensions-apiserver/pkg/apiserver/schema/defaulting"
	"k8s.io/kube-openapi/pkg/validation/strfmt"
	"k8s.io/kube-openapi/pkg/validation/validate"
	"sigs.k8s.io/yaml"

	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/model/rest/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/resources/registry"
	"github.com/kumahq/kuma/pkg/core/validators"
)

var YAML = &unmarshaler{unmarshalFn: func(bytes []byte, i interface{}) error {
	return yaml.Unmarshal(bytes, i)
}}
var JSON = &unmarshaler{unmarshalFn: json.Unmarshal}

type unmarshaler struct {
	unmarshalFn func([]byte, interface{}) error
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

func deepCopyMap(original map[string]interface{}) (map[string]interface{}, error) {
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(original); err != nil {
		return nil, fmt.Errorf("failed to encode original map: %w", err)
	}

	var copy map[string]interface{}
	if err := json.NewDecoder(&buf).Decode(&copy); err != nil {
		return nil, fmt.Errorf("failed to decode into copy map: %w", err)
	}

	return copy, nil
}

func printDifferences(original, modified map[string]interface{}) {
	if !reflect.DeepEqual(original, modified) {
		diff := cmp.Diff(original, modified)
		fmt.Printf("Differences between original and modified:\n%s\n", diff)
	} else {
		fmt.Println("No differences found between original and modified.")
	}
}

func (u *unmarshaler) Unmarshal(bytes []byte, desc core_model.ResourceTypeDescriptor) (Resource, error) {
	resource := desc.NewObject()
	restResource := From.Resource(resource)
	if desc.IsPluginOriginated {
		// desc.Schema is set only for new plugin originated policies
		rawObj := map[string]interface{}{}
		// Unfortunately to validate new policies we must first unmarshal into a rawObj
		if err := u.unmarshalFn(bytes, &rawObj); err != nil {
			return nil, &InvalidResourceError{Reason: fmt.Sprintf("invalid %s object: %q", desc.Name, err.Error())}
		}

		// Deep copy rawObj
		originalRawObj, err := deepCopyMap(rawObj)
		if err != nil {
			return nil, &InvalidResourceError{Reason: fmt.Sprintf("failed to deep copy rawObj: %q", err.Error())}
		}

		// Apply defaulting
		defaulting.Default(rawObj, desc.StructuralSchema)

		// Print differences
		printDifferences(originalRawObj, rawObj)

		validator := validate.NewSchemaValidator(desc.Schema, nil, "", strfmt.Default)
		res := validator.Validate(rawObj)
		if !res.IsValid() {
			return nil, toValidationError(res)
		}
	}

	if err := u.unmarshalFn(bytes, restResource); err != nil {
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
	for _, ri := range rsr.ResourceList.Items {
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
	rs.GetPagination().SetTotal(rsr.ResourceList.Total)
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
