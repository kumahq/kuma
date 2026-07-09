package labels

import (
	"fmt"
	"maps"
	"slices"
	"strings"

	"k8s.io/apimachinery/pkg/util/validation"

	config_core "github.com/kumahq/kuma/v3/pkg/config/core"
	core_model "github.com/kumahq/kuma/v3/pkg/core/resources/model"
)

type Violation struct {
	Key    string `json:"key"`
	Reason string `json:"reason"`
	Format bool   `json:"-"`
}

func (v Violation) String() string {
	if strings.Contains(v.Reason, v.Key) {
		return v.Reason
	}
	return "'" + v.Key + "' " + v.Reason
}

type ValidationContext struct {
	Mode                         config_core.CpMode
	Env                          config_core.EnvironmentType
	FederatedZone                bool
	ZoneName                     string
	Namespace                    Namespace
	DisableOriginLabelValidation bool
	Descriptor                   core_model.ResourceTypeDescriptor
	Spec                         core_model.ResourceSpec
	ResourceName                 string
	ResourceMesh                 string
	Privileged                   bool
}

type Result struct {
	Errors   []Violation
	Warnings []Violation
}

func mismatchReason(key, expected, got string) string {
	return fmt.Sprintf("%s is computed by the control plane (expected '%s'); the supplied value '%s' was overridden", key, expected, got)
}

func notApplicableReason(key, got string) string {
	return fmt.Sprintf("%s is managed by the control plane and is not applicable in this context; the supplied value '%s' was removed", key, got)
}

func systemOverriddenReason(key, got string) string {
	return fmt.Sprintf("%s is set by the control plane; the supplied value '%s' was overridden", key, got)
}

func Validate(labels map[string]string, ctx ValidationContext) Result {
	if ctx.Privileged {
		return Result{}
	}

	if format := formatViolations(labels); len(format) > 0 {
		return Result{Errors: format}
	}

	var errs, warns []Violation
	errs = append(errs, validateOrigin(labels, ctx)...)

	for _, key := range slices.Sorted(maps.Keys(registry)) {
		value, present := labels[key]
		if !present {
			continue
		}
		spec, ok := matchedSpec(registry[key], ctx)
		if !ok {
			warns = append(warns, Violation{Key: key, Reason: notApplicableReason(key, value)})
			continue
		}
		e, w := classifyLabel(spec, value, ctx)
		if e != nil {
			errs = append(errs, *e)
		}
		if w != nil {
			warns = append(warns, *w)
		}
	}

	return Result{Errors: errs, Warnings: warns}
}

func formatViolations(labels map[string]string) []Violation {
	var out []Violation
	for _, k := range slices.Sorted(maps.Keys(labels)) {
		for _, msg := range validation.IsQualifiedName(k) {
			out = append(out, Violation{Key: k, Reason: msg, Format: true})
		}
		for _, msg := range validation.IsValidLabelValue(labels[k]) {
			out = append(out, Violation{Key: k, Reason: msg, Format: true})
		}
	}
	return out
}

// classifyLabel checks a single present reserved label against its spec,
// returning an error and/or a warning violation (nil when there's nothing to report).
func classifyLabel(spec LabelSpec, value string, ctx ValidationContext) (*Violation, *Violation) {
	switch spec.Owner {
	case OwnerSystem:
		return nil, &Violation{Key: spec.Key, Reason: systemOverriddenReason(spec.Key, value)}

	case OwnerUser:
		if len(spec.AllowedValues) == 0 {
			return nil, nil
		}
		if slices.Contains(spec.AllowedValues, value) {
			return nil, nil
		}
		return &Violation{
			Key:    spec.Key,
			Reason: fmt.Sprintf("must be one of [%s]", strings.Join(quoted(spec.AllowedValues), ", ")),
		}, nil

	case OwnerControlPlane:
		expected, err := spec.Expected(ctx)
		if err != nil {
			// Compute surfaces malformed-resource errors on write.
			return nil, nil
		}
		if value != expected {
			return nil, &Violation{
				Key:    spec.Key,
				Reason: mismatchReason(spec.Key, expected, value),
			}
		}
		return nil, nil
	}

	return nil, nil
}

func quoted(vals []string) []string {
	out := make([]string, len(vals))
	for i, v := range vals {
		out[i] = "'" + v + "'"
	}
	return out
}
