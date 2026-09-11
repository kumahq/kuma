package mesh

import (
	"fmt"
	"slices"
	"strings"

	"github.com/spiffe/go-spiffe/v2/spiffeid"
	"google.golang.org/protobuf/proto"
	k8s_validation "k8s.io/apimachinery/pkg/util/validation"
	"sigs.k8s.io/yaml"

	common_api "github.com/kumahq/kuma/v3/api/common/v1alpha1"
	mesh_proto "github.com/kumahq/kuma/v3/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/v3/pkg/core/validators"
	"github.com/kumahq/kuma/v3/pkg/util/pointer"
	util_proto "github.com/kumahq/kuma/v3/pkg/util/proto"
)

type ValidateTargetRefOpts struct {
	SupportedKinds      []common_api.TargetRefKind
	SupportedKindsError string
	// AllowedInvalidNames is kept for compatibility with callers that still pass
	// legacy validation options while common TargetRef uses labels-only real
	// resource selectors.
	AllowedInvalidNames []string
	IsInboundPolicy     bool
	IsBackendRef        bool
}

func ValidateDuration(path validators.PathBuilder, duration *mesh_proto.Duration) validators.ValidationError {
	var errs validators.ValidationError
	if duration == nil {
		errs.AddViolationAt(path, "must have a positive value")
		return errs
	}
	if duration.AsDuration() == 0 {
		errs.AddViolationAt(path, "must have a positive value")
	}
	return errs
}

func ValidateThreshold(path validators.PathBuilder, threshold uint32) validators.ValidationError {
	var err validators.ValidationError
	if threshold == 0 {
		err.AddViolationAt(path, "must have a positive value")
	}
	return err
}

// ValidatePort validates that port is a valid TCP or UDP port number.
func ValidatePort(path validators.PathBuilder, port uint32) validators.ValidationError {
	err := validators.ValidationError{}

	if port == 0 || port > 65535 {
		err.AddViolationAt(path, "port must be in the range [1, 65535]")
	}

	return err
}

func AllowedValuesHint(values ...string) string {
	options := strings.Join(values, ", ")
	if len(values) == 0 {
		options = "(none)"
	}
	return fmt.Sprintf("Allowed values: %s", options)
}

// Resource is considered valid if it pass validation of any message
func ValidateAnyResourceYAML(resYAML string, msgs ...proto.Message) error {
	var err error
	for _, msg := range msgs {
		err = ValidateResourceYAML(msg, resYAML)
		if err == nil {
			return nil
		}
	}
	return err
}

// Resource is considered valid if it pass validation of any message
func ValidateAnyResourceYAMLPatch(resYAML string, msgs ...proto.Message) error {
	var err error
	for _, msg := range msgs {
		err = ValidateResourceYAMLPatch(msg, resYAML)
		if err == nil {
			return nil
		}
	}
	return err
}

func ValidateResourceYAML(msg proto.Message, resYAML string) error {
	json, err := yaml.YAMLToJSON([]byte(resYAML))
	if err != nil {
		json = []byte(resYAML)
	}

	if err := util_proto.FromJSON(json, msg); err != nil {
		return err
	}
	if v, ok := msg.(interface{ Validate() error }); ok {
		if err := v.Validate(); err != nil {
			return err
		}
	}
	return nil
}

func ValidateResourceYAMLPatch(msg proto.Message, resYAML string) error {
	json, err := yaml.YAMLToJSON([]byte(resYAML))
	if err != nil {
		json = []byte(resYAML)
	}
	return util_proto.FromJSON(json, msg)
}

func ValidateTargetRef(
	ref common_api.TargetRef,
	opts *ValidateTargetRefOpts,
) validators.ValidationError {
	var err validators.ValidationError

	if ref.Kind == "" {
		err.AddViolation("kind", validators.MustBeDefined)
		return err
	}
	if !slices.Contains(opts.SupportedKinds, ref.Kind) {
		errMsg := fmt.Sprintf("value '%s' is not supported", ref.Kind)
		if optsErr := opts.SupportedKindsError; optsErr != "" {
			errMsg = optsErr
		}
		err.AddViolation("kind", errMsg)
		return err
	}

	switch ref.Kind {
	case common_api.Mesh:
		err.Add(disallowedField("labels", pointer.Deref(ref.Labels), ref.Kind))
		err.Add(disallowedField("sectionName", pointer.Deref(ref.SectionName), ref.Kind))
	case common_api.Dataplane:
		if !opts.IsInboundPolicy && pointer.Deref(ref.SectionName) != "" {
			err.AddViolation("sectionName", "can only be used with inbound policies")
		}
	case common_api.MeshService:
		err.Add(requiredField("labels", pointer.Deref(ref.Labels), ref.Kind))
	case common_api.MeshHTTPRoute:
		err.Add(disallowedField("sectionName", pointer.Deref(ref.SectionName), ref.Kind))
		err.Add(requiredField("labels", pointer.Deref(ref.Labels), ref.Kind))
	case common_api.MeshExternalService:
		err.Add(disallowedField("sectionName", pointer.Deref(ref.SectionName), ref.Kind))
		err.Add(requiredField("labels", pointer.Deref(ref.Labels), ref.Kind))
	case common_api.MeshMultiZoneService:
		// sectionName selects a MeshMultiZoneService port and stays allowed,
		// mirroring MeshService and the pre-refactor behavior.
		err.Add(requiredField("labels", pointer.Deref(ref.Labels), ref.Kind))
	}

	return err
}

func ValidateMatch(match common_api.Match) validators.ValidationError {
	var verr validators.ValidationError
	if match.SpiffeID != nil {
		_, err := spiffeid.FromString(match.SpiffeID.Value)
		if err != nil {
			verr.AddViolation("spiffeID", fmt.Sprintf("must be a valid Spiffe ID: %s", err))
		}
	}
	if match.SNI != nil {
		switch match.SNI.Type {
		case common_api.SNIExactMatchType:
		case "":
			verr.AddViolation("sni.type", "must be set")
		default:
			verr.AddViolation("sni.type", fmt.Sprintf("unrecognized type %q, supported values are: Exact", match.SNI.Type))
		}
		if match.SNI.Value == "" {
			verr.AddViolation("sni.value", "must be set")
		} else {
			for _, violation := range k8s_validation.IsDNS1123Subdomain(match.SNI.Value) {
				verr.AddViolation("sni.value", violation)
			}
		}
	}
	return verr
}

func disallowedField[T ~string | ~map[string]string](
	name string,
	value T,
	kind common_api.TargetRefKind,
) validators.ValidationError {
	var err validators.ValidationError

	if isSet(value) {
		err.AddViolation(name, fmt.Sprintf("%s with kind %v", validators.MustNotBeSet, kind))
	}

	return err
}

func requiredField[T ~string | ~map[string]string](
	name string,
	value T,
	kind common_api.TargetRefKind,
) validators.ValidationError {
	var err validators.ValidationError

	if !isSet(value) {
		err.AddViolation(name, fmt.Sprintf("%s when kind is %v", validators.MustBeSet, kind))
	}

	return err
}

func isSet[T ~string | ~map[string]string](value T) bool {
	switch v := any(value).(type) {
	case string:
		return v != ""
	case map[string]string:
		return len(v) > 0
	default:
		return false
	}
}
