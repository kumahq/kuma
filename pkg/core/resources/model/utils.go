package model

import (
	"encoding/json"
	"path"
	"reflect"

	"google.golang.org/protobuf/types/known/anypb"
	"sigs.k8s.io/yaml"
)

func ToJSON(spec ResourceSpec) ([]byte, error) {
	return json.Marshal(spec)
}

func ToMap(spec ResourceSpec) (map[string]any, error) {
	v, err := ToJSON(spec)
	if err != nil {
		return nil, err
	}
	result := map[string]any{}
	if err := json.Unmarshal(v, &result); err != nil {
		return result, err
	}
	return result, nil
}

func ToYAML(spec ResourceSpec) ([]byte, error) {
	return yaml.Marshal(spec)
}

// ToAny writes the KDS wire form. Every spec goes as JSON with no type URL,
// which is the form policies have always been sent in.
func ToAny(spec ResourceSpec) (*anypb.Any, error) {
	bytes, err := json.Marshal(spec)
	if err != nil {
		return nil, err
	}
	return &anypb.Any{
		Value: bytes,
	}, nil
}

func FromJSON(src []byte, spec ResourceSpec) error {
	return json.Unmarshal(src, spec)
}

func FromYAML(src []byte, spec ResourceSpec) error {
	return yaml.Unmarshal(src, spec)
}

// FromAny reads what ToAny wrote. json.Unmarshal merges into what the target
// already holds, so a field the sender omitted would keep the value a previous
// read left behind, hence the reset.
func FromAny(src *anypb.Any, spec ResourceSpec) error {
	reset(spec)
	return json.Unmarshal(src.GetValue(), spec)
}

// reset returns a spec to its zero value, so that reading into it replaces
// rather than merges.
func reset(spec ResourceSpec) {
	v := reflect.ValueOf(spec)
	if v.Kind() != reflect.Pointer || v.IsNil() {
		return
	}
	v.Elem().Set(reflect.Zero(v.Elem().Type()))
}

func FullName(spec ResourceSpec) string {
	specType := reflect.TypeOf(spec).Elem()
	return path.Join(specType.PkgPath(), specType.Name())
}

func Equal(x, y ResourceSpec) bool {
	return reflect.DeepEqual(x, y)
}

func IsEmpty(spec ResourceSpec) bool {
	return reflect.ValueOf(spec).Elem().IsZero()
}
