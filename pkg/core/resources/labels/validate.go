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

// Violation is a single label-validation finding.
//
// Format=true marks failures of the K8s label format constraints
// (qualified-name keys, label-value values). Callers may treat these
// differently from semantic findings — the K8s webhook short-circuits on
// format violations so the K8s API server's native "Invalid value: ..." error
// surfaces unmodified.
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

// ValidationContext carries the per-call information validators consult.
// It is the only piece of state passed into Validate; all per-label rules
// derive their decision from these fields.
type ValidationContext struct {
	Mode                         config_core.CpMode
	IsK8s                        bool
	FederatedZone                bool
	ZoneName                     string
	Namespace                    Namespace
	DisableOriginLabelValidation bool
	Descriptor                   core_model.ResourceTypeDescriptor
	Spec                         core_model.ResourceSpec
	ResourceName                 string
	ResourceMesh                 string
	// Privileged: when true, Validate returns an empty Result with the input
	// labels echoed in Sanitized. Set by callers that bypass validation (KDS
	// sync, GC, storage-version migrator).
	Privileged bool
}

// Result is what Validate returns.
//
//   - Errors must reject the request (format issues, OwnerUser AllowedValues
//     mismatch, origin vocabulary/mismatch/required-presence).
//   - Warnings should be surfaced to the caller without rejecting. They cover
//     attempts to set CP-managed labels: Compute will override (or drop) the
//     value and the user should know what happened.
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

// Validate runs format + semantic checks against the given label set.
//
// Format violations short-circuit the rest of the function (validators
// assume well-formed input). All format findings have Format=true; callers
// that need to behave differently for format issues inspect Violation.Format.
//
// kuma.io/origin is checked first by validateOrigin — its semantics are
// stricter (vocabulary errors, CP-vs-user mismatch as error) and don't share
// the warning/override pattern used by the rest of the CP-owned labels.
//
// Semantic classification for the registry (everything except origin), per
// LabelSpec.Owner:
//
//   - OwnerSystem        — any user-set value is a warning. Compute will
//     overwrite or drop it.
//   - OwnerUser          — any value is accepted; if AllowedValues is non-empty
//     and the value is outside the set, that is an error.
//   - OwnerControlPlane  — Expected(ctx) supplies the CP value. The user value
//     is accepted only when it matches expected (or OpenValue is set). All
//     other cases are warnings; Compute will regenerate the right value.
//
// Reserved keys (kuma.io/* or k8s.kuma.io/*) that are not in the registry are
// left alone — Compute will not touch them either, so they pass through as
// opaque user-set values.
//
// When ctx.Privileged is true the function returns an empty Result — KDS-synced
// and CP-internal flows skip all validation.
func Validate(labels map[string]string, ctx ValidationContext) Result {
	if ctx.Privileged {
		return Result{}
	}

	if format := formatViolations(labels); len(format) > 0 {
		return Result{Errors: format}
	}

	var errs, warns []Violation
	errs = append(errs, validateOrigin(labels, ctx)...)

	for _, spec := range registry {
		value, present := labels[spec.Key]
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

// classifyOne returns (error, warning) for a single registered label.
// Exactly one is non-nil, or both are nil when the value is accepted.
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
		var expected string
		applies := true
		if spec.Expected != nil {
			expected, applies = spec.Expected(ctx)
		}

		if !applies {
			if !present {
				return nil, nil
			}
			return nil, &Violation{Key: spec.Key, Reason: notApplicableReason(spec.Key, value)}
		}

		if present {
			if spec.OpenValue {
				return nil, nil
			}
			if value != expected {
				return nil, &Violation{
					Key:    spec.Key,
					Reason: mismatchReason(spec.Key, expected, value),
				}
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
