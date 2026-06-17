package labels

import (
	"fmt"
	"slices"
	"sort"
	"strings"

	"k8s.io/apimachinery/pkg/util/validation"

	config_core "github.com/kumahq/kuma/v2/pkg/config/core"
	core_model "github.com/kumahq/kuma/v2/pkg/core/resources/model"
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

	for key, specs := range registry {
		value, present := labels[key]
		spec, ok := matchedSpec(specs, ctx)
		if !ok {
			if present {
				warns = append(warns, Violation{Key: key, Reason: notApplicableReason(key, value)})
			}
			continue
		}
		e, w := classifyOne(spec, value, present, ctx)
		if e != nil {
			errs = append(errs, *e)
		}
		if w != nil {
			warns = append(warns, *w)
		}
	}

	sort.Slice(errs, func(i, j int) bool { return errs[i].Key < errs[j].Key })
	sort.Slice(warns, func(i, j int) bool { return warns[i].Key < warns[j].Key })
	return Result{Errors: errs, Warnings: warns}
}

func formatViolations(labels map[string]string) []Violation {
	var out []Violation
	for _, k := range sortedKeys(labels) {
		for _, msg := range validation.IsQualifiedName(k) {
			out = append(out, Violation{Key: k, Reason: msg, Format: true})
		}
		for _, msg := range validation.IsValidLabelValue(labels[k]) {
			out = append(out, Violation{Key: k, Reason: msg, Format: true})
		}
	}
	return out
}

func sortedKeys(m map[string]string) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func classifyOne(spec LabelSpec, value string, present bool, ctx ValidationContext) (*Violation, *Violation) {
	switch spec.Owner {
	case OwnerSystem:
		if present {
			return nil, &Violation{Key: spec.Key, Reason: systemOverriddenReason(spec.Key, value)}
		}
		return nil, nil

	case OwnerUser:
		if !present {
			return nil, nil
		}
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
		if !present {
			return nil, nil
		}
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
