package mesh

import (
	"fmt"
	"sort"

	mesh_proto "github.com/Kong/kuma/api/mesh/v1alpha1"
	"github.com/Kong/kuma/pkg/core/validators"
)

type SelectorValidatorFunc func(path validators.PathBuilder, selector map[string]string) validators.ValidationError
type TagKeyValidatorFunc func(path validators.PathBuilder, key string) validators.ValidationError
type TagValueValidatorFunc func(path validators.PathBuilder, key, value string) validators.ValidationError

type ValidateSelectorOpts struct {
	SkipRequireAtLeastOneTag bool
	ExtraSelectorValidators  []SelectorValidatorFunc
	ExtraTagKeyValidators    []TagKeyValidatorFunc
	ExtraTagValueValidators  []TagValueValidatorFunc
}

type ValidateSelectorsOpts struct {
	ValidateSelectorOpts
	RequireAtLeastOneSelector bool
}

func ValidateSelectors(path validators.PathBuilder, sources []*mesh_proto.Selector, opts ValidateSelectorsOpts) (err validators.ValidationError) {
	if opts.RequireAtLeastOneSelector && len(sources) == 0 {
		err.AddViolationAt(path, "must have at least one element")
	}
	for i, selector := range sources {
		err.Add(ValidateSelector(path.Index(i).Field("match"), selector.GetMatch(), opts.ValidateSelectorOpts))
	}
	return
}

func ValidateSelector(path validators.PathBuilder, selector map[string]string, opts ValidateSelectorOpts) (err validators.ValidationError) {
	if !opts.SkipRequireAtLeastOneTag && len(selector) == 0 {
		err.AddViolationAt(path, "must have at least one tag")
	}
	for _, validate := range opts.ExtraSelectorValidators {
		err.Add(validate(path, selector))
	}
	for _, key := range Keys(selector) {
		if key == "" {
			err.AddViolationAt(path, "tag key must be non-empty")
		}
		for _, validate := range opts.ExtraTagKeyValidators {
			err.Add(validate(path, key))
		}
		value := selector[key]
		if value == "" {
			err.AddViolationAt(path.Key(key), "tag value must be non-empty")
		}
		for _, validate := range opts.ExtraTagValueValidators {
			err.Add(validate(path, key, value))
		}
	}
	_, defined := selector[mesh_proto.ServiceTag]
	if !defined {
		err.AddViolationAt(path, fmt.Sprintf("mandatory tag %q is missing", mesh_proto.ServiceTag))
	}
	return
}

var OnlyServiceTagAllowed = ValidateSelectorsOpts{
	RequireAtLeastOneSelector: true,
	ValidateSelectorOpts: ValidateSelectorOpts{
		SkipRequireAtLeastOneTag: true,
		ExtraSelectorValidators: []SelectorValidatorFunc{
			func(path validators.PathBuilder, selector map[string]string) (err validators.ValidationError) {
				_, defined := selector[mesh_proto.ServiceTag]
				if len(selector) != 1 || !defined {
					err.AddViolationAt(path, fmt.Sprintf("must consist of exactly one tag %q", mesh_proto.ServiceTag))
				}
				return
			},
		},
		ExtraTagKeyValidators: []TagKeyValidatorFunc{
			func(path validators.PathBuilder, key string) (err validators.ValidationError) {
				if key != mesh_proto.ServiceTag {
					err.AddViolationAt(path.Key(key), fmt.Sprintf("tag %q is not allowed", key))
				}
				return
			},
		},
	},
}

func Keys(tags map[string]string) []string {
	// sort keys for consistency
	var keys []string
	for key := range tags {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}
