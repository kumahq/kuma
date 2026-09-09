package v1alpha1

import (
	"encoding/json"
	"strconv"

	"github.com/evanphx/json-patch/v5"
)

// JsonPatchBlock is one json patch operation block.
type JsonPatchBlock struct {
	// Op is a jsonpatch operation string.
	// +required
	// +kubebuilder:validation:Enum=add;remove;replace;move;copy
	Op string `json:"op"`
	// Path is a jsonpatch path string.
	// +required
	Path string `json:"path"`
	// Value must be a valid json value used by replace and add operations.
	// +kubebuilder:validation:Schemaless
	// +kubebuilder:pruning:PreserveUnknownFields
	// +kuma:nolint // json.RawMessage is already nilable, so an unset value stays distinguishable from an empty one without wrapping it in a pointer
	Value json.RawMessage `json:"value,omitempty"`
	// From is a jsonpatch from string, used by move and copy operations.
	From *string `json:"from,omitempty"`
}

func ToJsonPatch(in []JsonPatchBlock) jsonpatch.Patch {
	var res []jsonpatch.Operation

	for _, o := range in {
		var fromString string
		if o.From != nil {
			fromString = *o.From
		}

		op := json.RawMessage(strconv.Quote(o.Op))
		from := json.RawMessage(strconv.Quote(fromString))
		path := json.RawMessage(strconv.Quote(o.Path))
		value := o.Value

		res = append(res, jsonpatch.Operation{
			"op":    &op,
			"path":  &path,
			"from":  &from,
			"value": &value,
		})
	}

	return res
}
