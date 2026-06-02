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

// Violation is a single label-validation failure.
//
// Format=true marks failures of the K8s label format constraints
// (qualified-name keys, label-value values). Callers may treat these
// differently from semantic violations — the REST API emits both, the K8s
// webhook short-circuits on format violations so the K8s API server's native
// "Invalid value: ..." error surfaces unmodified.
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
	SystemNamespace              string
	Namespace                    Namespace
	DisableOriginLabelValidation bool
	Descriptor                   core_model.ResourceTypeDescriptor
	Spec                         core_model.ResourceSpec
	ResourceName                 string
	ResourceMesh                 string
	// Privileged: when true, Validate returns nil. Set by callers that bypass
	// validation (KDS sync, GC, storage-version migrator).
	Privileged bool
}

const reservedReason = "is a reserved label managed by the control plane and cannot be set on apply"

// Validate runs format + semantic checks against the given label set.
//
// Format violations are returned first; when any are present, semantic
// validation is skipped (validators assume well-formed input). All returned
// violations have Format=true; callers that need to behave differently for
// format issues (e.g. the K8s webhook, which lets the API server emit the
// native format error) inspect Violation.Format.
//
// Semantic behavior, per LabelSpec.Owner:
//   - OwnerSystem        — any user-set value is rejected.
//   - OwnerUser          — any value is accepted; if AllowedValues is non-empty,
//     the value must be in the set.
//   - OwnerControlPlane  — Expected(ctx) supplies the CP value. If applies=false,
//     any user-set value is rejected. Otherwise the user
//     value (if present) must match expected. If absent and
//     RequirePresence(ctx) is true, that is also a violation.
//
// Reserved keys (kuma.io/* or k8s.kuma.io/*) absent from the registry are
// rejected as reserved.
//
// When ctx.Privileged is true the function returns nil — KDS-synced and
// CP-internal flows skip all validation.
func Validate(labels map[string]string, ctx ValidationContext) []Violation {
	if ctx.Privileged {
		return nil
	}

	if format := formatViolations(labels); len(format) > 0 {
		return format
	}

	var violations []Violation

	for _, spec := range registry {
		value, present := labels[spec.Key]
		if v := validateOne(spec, value, present, ctx); v != nil {
			violations = append(violations, *v)
		}
	}

	for key := range labels {
		if _, known := registry[key]; known {
			continue
		}
		if !mesh_proto.IsReservedLabelKey(key) {
			continue
		}
		violations = append(violations, Violation{Key: key, Reason: reservedReason})
	}

	sort.Slice(violations, func(i, j int) bool { return violations[i].Key < violations[j].Key })
	return violations
}

// ValidateOrigin runs only the kuma.io/origin label spec.
// Used by the delete flow, where the other labels are CP-managed state and
// the only thing the caller can authoritatively gate on is origin.
func ValidateOrigin(labels map[string]string, ctx ValidationContext) []Violation {
	if ctx.Privileged {
		return nil
	}
	spec, ok := registry[mesh_proto.ResourceOriginLabel]
	if !ok {
		return nil
	}
	value, present := labels[spec.Key]
	if v := validateOne(spec, value, present, ctx); v != nil {
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

func validateOne(spec LabelSpec, value string, present bool, ctx ValidationContext) *Violation {
	switch spec.Owner {
	case OwnerSystem:
		if present {
			return &Violation{Key: spec.Key, Reason: reservedReason}
		}
		return nil

	case OwnerUser:
		if !present {
			return nil
		}
		if len(spec.AllowedValues) == 0 {
			return nil
		}
		if slices.Contains(spec.AllowedValues, value) {
			return nil
		}
		return &Violation{
			Key:    spec.Key,
			Reason: fmt.Sprintf("must be one of [%s]", strings.Join(quoted(spec.AllowedValues), ", ")),
		}

	case OwnerControlPlane:
		var expected string
		applies := true
		if spec.Expected != nil {
			expected, applies = spec.Expected(ctx)
		}

		if !applies {
			if present {
				return &Violation{Key: spec.Key, Reason: reservedReason}
			}
			return nil
		}

		if present {
			if spec.OpenValue {
				return nil
			}
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

	return nil
}

func quoted(vals []string) []string {
	out := make([]string, len(vals))
	for i, v := range vals {
		out[i] = "'" + v + "'"
	}
	return out
}
