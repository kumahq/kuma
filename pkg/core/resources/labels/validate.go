package labels

import (
	"fmt"
	"slices"
	"sort"
	"strings"

	"k8s.io/apimachinery/pkg/util/validation"

	mesh_proto "github.com/kumahq/kuma/v2/api/mesh/v1alpha1"
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
//     mismatch, OwnerControlPlane StrictMatch mismatch).
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

func strictMismatchReason(key, expected, got string) string {
	return fmt.Sprintf("%s should be '%s', got '%s'", key, expected, got)
}

func notApplicableReason(key, got string) string {
	return fmt.Sprintf("%s is managed by the control plane and is not applicable in this context; the supplied value '%s' was removed", key, got)
}

func systemOverriddenReason(key, got string) string {
	return fmt.Sprintf("%s is set by the control plane; the supplied value '%s' was overridden", key, got)
}

func notAllowedValueReason(key string, allowed []string, got string) string {
	return fmt.Sprintf("%s must be one of [%s] in this context; the supplied value '%s' was removed", key, strings.Join(quoted(allowed), ", "), got)
}

// Validate runs format + semantic checks against the given label set.
//
// Format violations short-circuit the rest of the function (validators
// assume well-formed input). All format findings have Format=true; callers
// that need to behave differently for format issues inspect Violation.Format.
//
// Semantic classification, per LabelSpec.Owner:
//
//   - OwnerSystem        — any user-set value is a warning. Compute will
//     overwrite or drop it.
//   - OwnerUser          — any value is accepted; if AllowedValues is non-empty
//     and the value is outside the set, that is an error.
//   - OwnerControlPlane  — Expected(ctx) supplies the CP value. The user value
//     is accepted only when it matches expected (or OpenValue is set, or
//     applies=false with a valid AllowAnyWhenNotApplicable value). All other
//     cases are warnings; Compute will regenerate the right value. Exception:
//     when StrictMatch is set on the spec, a mismatch against an applicable
//     Expected becomes an error (kuma.io/origin).
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

// ValidateOriginFormat is a context-free vocabulary check on the
// kuma.io/origin label — its value must be one of "global" or "zone" (or
// empty, treated as unset). Appropriate for flows that intentionally skip
// the context-aware ValidateOrigin (non-federated zone CPs, legacy
// non-plugin-originated resources) but still want to reject unknown values.
func ValidateOriginFormat(labels map[string]string) []Violation {
	value, ok := labels[mesh_proto.ResourceOriginLabel]
	if !ok || value == "" {
		return nil
	}
	if err := mesh_proto.ResourceOrigin(value).IsValid(); err != nil {
		return []Violation{{
			Key:    mesh_proto.ResourceOriginLabel,
			Reason: fmt.Sprintf("%s should be 'global' or 'zone', got '%s'", mesh_proto.ResourceOriginLabel, value),
		}}
	}
	return nil
}

// ValidateOrigin runs only the kuma.io/origin label spec, returning errors
// only. Used by the delete flow — it is the one CP-computed label the
// delete-authorization gate actually needs to enforce. A mismatch here is
// "you are deleting in the wrong CP", not a value the CP would silently
// override.
func ValidateOrigin(labels map[string]string, ctx ValidationContext) []Violation {
	if ctx.Privileged {
		return nil
	}
	spec, ok := registry[mesh_proto.ResourceOriginLabel]
	if !ok {
		return nil
	}
	value, present := labels[spec.Key]
	if v := originDeleteCheck(spec, value, present, ctx); v != nil {
		return []Violation{*v}
	}
	return nil
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
			if spec.AllowAnyWhenNotApplicable {
				if len(spec.AllowedValues) > 0 && !slices.Contains(spec.AllowedValues, value) {
					// Out-of-vocabulary value — the CP would not have set it.
					// Strip it and warn.
					return nil, &Violation{
						Key:    spec.Key,
						Reason: notAllowedValueReason(spec.Key, spec.AllowedValues, value),
					}
				}
				return nil, nil
			}
			return nil, &Violation{Key: spec.Key, Reason: notApplicableReason(spec.Key, value)}
		}

		if present {
			if spec.OpenValue {
				return nil, nil
			}
			if value != expected {
				if spec.StrictMatch {
					return &Violation{
						Key:    spec.Key,
						Reason: strictMismatchReason(spec.Key, expected, value),
					}, nil
				}
				return nil, &Violation{
					Key:    spec.Key,
					Reason: mismatchReason(spec.Key, expected, value),
				}
			}
			return nil, nil
		}

		// Absent and applies=true. Most CP-owned labels are populated by Compute,
		// so we don't fail here. The exception is RequirePresence: the user must
		// consciously set the value (kuma.io/origin: zone on a zone CP gates apply
		// so users don't blindly create resources thinking they're on Global).
		if spec.RequirePresence != nil && spec.RequirePresence(ctx) {
			return &Violation{
				Key:    spec.Key,
				Reason: fmt.Sprintf("the %s label must be set to '%s'", spec.Key, expected),
			}, nil
		}
		return nil, nil
	}

	return nil, nil
}

// originDeleteCheck is the validator used by the delete flow — it preserves
// pre-warning behavior so origin mismatches still reject deletes.
func originDeleteCheck(spec LabelSpec, value string, present bool, ctx ValidationContext) *Violation {
	var expected string
	applies := true
	if spec.Expected != nil {
		expected, applies = spec.Expected(ctx)
	}

	if !applies {
		if !present {
			return nil
		}
		if spec.AllowAnyWhenNotApplicable {
			if len(spec.AllowedValues) > 0 && !slices.Contains(spec.AllowedValues, value) {
				return &Violation{
					Key:    spec.Key,
					Reason: fmt.Sprintf("must be one of [%s]", strings.Join(quoted(spec.AllowedValues), ", ")),
				}
			}
			return nil
		}
		return &Violation{Key: spec.Key, Reason: "is a reserved label managed by the control plane and cannot be set on apply"}
	}

	if present {
		if value != expected {
			return &Violation{
				Key:    spec.Key,
				Reason: fmt.Sprintf("%s should be '%s', got '%s'", spec.Key, expected, value),
			}
		}
		return nil
	}

	if spec.RequirePresence != nil && spec.RequirePresence(ctx) {
		return &Violation{
			Key:    spec.Key,
			Reason: fmt.Sprintf("the %s label must be set to '%s'", spec.Key, expected),
		}
	}
	return nil
}

func quoted(vals []string) []string {
	out := make([]string, len(vals))
	for i, v := range vals {
		out[i] = "'" + v + "'"
	}
	return out
}
