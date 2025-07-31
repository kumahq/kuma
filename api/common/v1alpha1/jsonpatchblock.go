package v1alpha1

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/evanphx/json-patch/v5"
	"go.uber.org/multierr"
)

// JsonPatchBlock is one json patch operation block.
type JsonPatchBlock struct {
	// Op is a jsonpatch operation string.
	// +required
	// +kubebuilder:validation:Enum=add;remove;replace;move;copy
	Op string `json:"op"`
	// Path is a jsonpatch path string.
	// +required
	// +kuma:nolint
	Path *string `json:"path"`
	// Value must be a valid json value used by replace and add operations.
	// +kubebuilder:validation:Schemaless
	// +kubebuilder:pruning:PreserveUnknownFields
	// +kuma:nolint
	Value json.RawMessage `json:"value,omitempty"`
	// From is a jsonpatch from string, used by move and copy operations.
	From *string `json:"from,omitempty"`
}

func ToJsonPatch(in []JsonPatchBlock) (jsonpatch.Patch, error) {
	var errs error
	var res []jsonpatch.Operation

	for _, o := range in {
		if o.Path == nil {
			errs = multierr.Append(errs, fmt.Errorf("path must be defined"))
			continue
		}

		var fromString string
		if o.From != nil {
			fromString = *o.From
		}

		op := json.RawMessage(strconv.Quote(o.Op))
		from := json.RawMessage(strconv.Quote(fromString))
		path := json.RawMessage(strconv.Quote(*o.Path))
		value := o.Value

		res = append(res, jsonpatch.Operation{
			"op":    &op,
			"path":  &path,
			"from":  &from,
			"value": &value,
		})
	}

	return res, errs
}
