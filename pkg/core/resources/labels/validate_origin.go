package labels

import (
	"fmt"

	mesh_proto "github.com/kumahq/kuma/v2/api/mesh/v1alpha1"
	config_core "github.com/kumahq/kuma/v2/pkg/config/core"
)

func validateOrigin(labels map[string]string, ctx ValidationContext) []Violation {
	value, present := labels[mesh_proto.ResourceOriginLabel]

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

func expectedOrigin(ctx ValidationContext) string {
	if ctx.Mode == config_core.Global {
		return string(mesh_proto.GlobalResourceOrigin)
	}
	return string(mesh_proto.ZoneResourceOrigin)
}

func requireOriginPresence(ctx ValidationContext) bool {
	if ctx.Mode != config_core.Zone {
		return false
	}
	if ctx.Env == config_core.KubernetesEnvironment {
		return ctx.Namespace.system
	}
	return ctx.FederatedZone && ctx.Descriptor.IsPluginOriginated
}

// ValidateOriginFormat checks the kuma.io/origin vocabulary only.
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

// ValidateOrigin runs origin checks for delete authorization.
func ValidateOrigin(labels map[string]string, ctx ValidationContext) []Violation {
	if ctx.Privileged {
		return nil
	}
	return validateOrigin(labels, ctx)
}
