package labels

import (
	"fmt"

	mesh_proto "github.com/kumahq/kuma/v2/api/mesh/v1alpha1"
	config_core "github.com/kumahq/kuma/v2/pkg/config/core"
)

// kuma.io/origin sits outside the LabelSpec registry. Its validation rules
// don't share the warn/override pattern the registry models:
//
//   - Out-of-vocabulary values are errors (not warnings) even when the broader
//     CP-vs-user check is disabled — vocabulary is independent of the disable
//     flag.
//   - A CP-vs-user mismatch is an error (not a warning) because it signals
//     "you are creating/deleting on the wrong CP" rather than something Compute
//     should silently fix.
//   - Some contexts require the user to consciously set kuma.io/origin so they
//     don't blindly create resources thinking they're on Global.
//
// validateOrigin owns all three checks. Compute calls expectedOrigin directly
// to know what value to write (or to delete the label when not applicable).

// validateOrigin returns origin-related errors, ordered: vocabulary first
// (always runs, even with DisableOriginLabelValidation), then CP-vs-user
// match and required-presence (both skipped when validation is disabled).
//
// The same logic serves apply and delete paths — delete authorization only
// hinges on "is the user touching a resource that belongs to this CP".
func validateOrigin(labels map[string]string, ctx ValidationContext) []Violation {
	value, present := labels[mesh_proto.ResourceOriginLabel]

	// Vocabulary check runs unconditionally. DisableOriginLabelValidation
	// turns off the CP-vs-user comparison; it does not relax what counts as
	// a valid origin string.
	if present && value != "" {
		if err := mesh_proto.ResourceOrigin(value).IsValid(); err != nil {
			return []Violation{{
				Key:    mesh_proto.ResourceOriginLabel,
				Reason: fmt.Sprintf("%s should be 'global' or 'zone', got '%s'", mesh_proto.ResourceOriginLabel, value),
			}}
		}
	}

	if ctx.DisableOriginLabelValidation {
		return nil
	}

	expected := expectedOrigin(ctx)

	if present {
		if value != expected {
			return []Violation{{
				Key:    mesh_proto.ResourceOriginLabel,
				Reason: fmt.Sprintf("%s should be '%s', got '%s'", mesh_proto.ResourceOriginLabel, expected, value),
			}}
		}
		return nil
	}

	if requireOriginPresence(ctx) {
		return []Violation{{
			Key:    mesh_proto.ResourceOriginLabel,
			Reason: fmt.Sprintf("the %s label must be set to '%s'", mesh_proto.ResourceOriginLabel, expected),
		}}
	}
	return nil
}

// expectedOrigin reports the CP-computed kuma.io/origin value for ctx.
// Origin is deterministic from the CP mode: anything that reaches us on
// Global was authored there ("global"); anything that reaches us on Zone
// was authored there ("zone"). KDS-synced resources skip validation
// entirely via the Privileged bypass, so this function never has to model
// a "not authored here" case.
func expectedOrigin(ctx ValidationContext) string {
	if ctx.Mode == config_core.Global {
		return string(mesh_proto.GlobalResourceOrigin)
	}
	return string(mesh_proto.ZoneResourceOrigin)
}

// requireOriginPresence reports whether the user MUST set kuma.io/origin
// explicitly on apply. The "conscious-apply gate" prevents users from
// blindly creating resources on a zone thinking they're on Global.
func requireOriginPresence(ctx ValidationContext) bool {
	if ctx.Mode != config_core.Zone {
		return false
	}
	if ctx.IsK8s {
		return ctx.Namespace.system
	}
	return ctx.FederatedZone && ctx.Descriptor.IsPluginOriginated
}

// ValidateOriginFormat is a context-free vocabulary check on the
// kuma.io/origin label — its value must be one of "global" or "zone" (or
// empty, treated as unset). Used by webhook paths that intentionally skip
// the context-aware validation (non-federated zone CPs, legacy
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

// ValidateOrigin runs origin-specific checks for the delete authorization
// flow. A mismatch here is "you are deleting from the wrong CP" — kept as
// an error so we don't silently delete from the wrong place.
func ValidateOrigin(labels map[string]string, ctx ValidationContext) []Violation {
	if ctx.Privileged {
		return nil
	}
	return validateOrigin(labels, ctx)
}
