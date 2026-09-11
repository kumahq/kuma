package model_test

import (
	"encoding/json"
	"testing"

	"google.golang.org/protobuf/types/known/anypb"

	core_model "github.com/kumahq/kuma/v3/pkg/core/resources/model"
	meshtimeout_api "github.com/kumahq/kuma/v3/pkg/plugins/policies/meshtimeout/api/v1alpha1"
)

// Policy specs are Go structs rather than protobuf messages, so they travel the
// KDS wire as type-URL-less JSON and are read back with json.Unmarshal. That
// merges into the target, which would keep a field the sender omitted at
// whatever a previous read left behind.
func TestFromAnyReplacesRatherThanMerges(t *testing.T) {
	spec := &meshtimeout_api.MeshTimeout{}
	if err := json.Unmarshal([]byte(`{"targetRef":{"kind":"Mesh"}}`), spec); err != nil {
		t.Fatal(err)
	}
	if spec.TargetRef == nil {
		t.Fatal("seeding the target failed")
	}

	empty, err := json.Marshal(&meshtimeout_api.MeshTimeout{})
	if err != nil {
		t.Fatal(err)
	}
	if err := core_model.FromAny(&anypb.Any{Value: empty}, spec); err != nil {
		t.Fatal(err)
	}

	if spec.TargetRef != nil {
		t.Fatalf("target kept a value the sender omitted: %+v", spec.TargetRef)
	}
}
