package mesh

import (
	"fmt"
	"regexp"
	"slices"
	"sort"
	"strings"

	"github.com/spiffe/go-spiffe/v2/spiffeid"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/durationpb"
	k8s_validation "k8s.io/apimachinery/pkg/util/validation"
	"sigs.k8s.io/yaml"

	common_api "github.com/kumahq/kuma/v3/api/common/v1alpha1"
	mesh_proto "github.com/kumahq/kuma/v3/api/mesh/v1alpha1"
	core_meta "github.com/kumahq/kuma/v3/pkg/core/metadata"
	"github.com/kumahq/kuma/v3/pkg/core/validators"
	"github.com/kumahq/kuma/v3/pkg/util/pointer"
	util_proto "github.com/kumahq/kuma/v3/pkg/util/proto"
)

const dnsLabel = `[a-z0-9]([-a-z0-9]*[a-z0-9])?`

var (
	DomainRegexp         = regexp.MustCompile("^" + dnsLabel + "(\\." + dnsLabel + ")*" + "$")
	tagNameCharacterSet  = regexp.MustCompile(`^[a-zA-Z0-9.\-_:/]*$`)
	tagValueCharacterSet = regexp.MustCompile(`^[a-zA-Z0-9.\-_:]*$`)
	selectorCharacterSet = regexp.MustCompile(`^([a-zA-Z0-9.\-_:/]*|\*)$`)
)

type (
	TagsValidatorFunc     func(path validators.PathBuilder, selector map[string]string) validators.ValidationError
	TagKeyValidatorFunc   func(path validators.PathBuilder, key string) validators.ValidationError
	TagValueValidatorFunc func(path validators.PathBuilder, key, value string) validators.ValidationError
)

type ValidateTagsOpts struct {
	RequireAtLeastOneTag    bool
	RequireService          bool
	ForbidService           bool
	ExtraTagsValidators     []TagsValidatorFunc
	ExtraTagKeyValidators   []TagKeyValidatorFunc
	ExtraTagValueValidators []TagValueValidatorFunc
}

type ValidateSelectorsOpts struct {
	ValidateTagsOpts
	RequireAtMostOneSelector  bool
	RequireAtLeastOneSelector bool
}

type ValidateTargetRefOpts struct {
	SupportedKinds             []common_api.TargetRefKind
	SupportedKindsError        string
	GatewayListenerTagsAllowed bool
	// AllowedInvalidNames is kept for compatibility with callers that still pass
	// legacy validation options while common TargetRef uses labels-only real
	// resource selectors.
	AllowedInvalidNames []string
	IsInboundPolicy     bool
	IsBackendRef        bool
}

func ValidateSelectors(path validators.PathBuilder, sources []*mesh_proto.Selector, opts ValidateSelectorsOpts) validators.ValidationError {
	var err validators.ValidationError
	if opts.RequireAtLeastOneSelector && len(sources) == 0 {
		err.AddViolationAt(path, "must have at least one element")
	}

	for i, selector := range sources {
		err.Add(ValidateSelector(path.Index(i).Field("match"), selector.GetMatch(), opts.ValidateTagsOpts))
		if i > 0 && opts.RequireAtMostOneSelector {
			err.AddViolationAt(path.Index(i), `there can be at most one selector`)
		}
	}
	return err
}

func ValidateSelector(path validators.PathBuilder, tags map[string]string, opts ValidateTagsOpts) validators.ValidationError {
	opts.ExtraTagValueValidators = append([]TagValueValidatorFunc{
		func(path validators.PathBuilder, key, value string) validators.ValidationError {
			var err validators.ValidationError
			if !selectorCharacterSet.MatchString(value) {
				err.AddViolationAt(path.Key(key), `tag value must consist of alphanumeric characters, dots, dashes, slashes and underscores or be "*"`)
			}
			return err
		},
	}, opts.ExtraTagValueValidators...)

	return validateTagKeyValues(path, tags, opts)
}

func ValidateTags(path validators.PathBuilder, tags map[string]string, opts ValidateTagsOpts) validators.ValidationError {
	opts.ExtraTagValueValidators = append([]TagValueValidatorFunc{
		func(path validators.PathBuilder, key, value string) validators.ValidationError {
			var err validators.ValidationError
			if !tagValueCharacterSet.MatchString(value) {
				err.AddViolationAt(path.Key(key), "tag value must consist of alphanumeric characters, dots, dashes and underscores")
			}
			return err
		},
	}, opts.ExtraTagValueValidators...)

	return validateTagKeyValues(path, tags, opts)
}

func validateTagKeyValues(path validators.PathBuilder, keyValues map[string]string, opts ValidateTagsOpts) validators.ValidationError {
	var err validators.ValidationError
	if opts.RequireAtLeastOneTag && len(keyValues) == 0 {
		err.AddViolationAt(path, "must have at least one tag")
	}
	for _, validate := range opts.ExtraTagsValidators {
		err.Add(validate(path, keyValues))
	}
	for _, key := range Keys(keyValues) {
		if key == "" {
			err.AddViolationAt(path, "tag name must be non-empty")
		}
		if !tagNameCharacterSet.MatchString(key) {
			err.AddViolationAt(path.Key(key), "tag name must consist of alphanumeric characters, dots, dashes, slashes and underscores")
		}
		for _, validate := range opts.ExtraTagKeyValidators {
			err.Add(validate(path, key))
		}

		value := keyValues[key]
		if value == "" {
			err.AddViolationAt(path.Key(key), "tag value must be non-empty")
		}
		for _, validate := range opts.ExtraTagValueValidators {
			err.Add(validate(path, key, value))
		}
	}
	_, defined := keyValues[mesh_proto.ServiceTag]
	if opts.ForbidService && defined {
		err.AddViolationAt(path, fmt.Sprintf("%q must not be defined", mesh_proto.ServiceTag))
	}
	if opts.RequireService && !defined {
		err.AddViolationAt(path, fmt.Sprintf("mandatory tag %q is missing", mesh_proto.ServiceTag))
	}
	return err
}

var OnlyServiceTagAllowed = ValidateSelectorsOpts{
	RequireAtLeastOneSelector: true,
	ValidateTagsOpts: ValidateTagsOpts{
		RequireService: true,
		ExtraTagsValidators: []TagsValidatorFunc{
			func(path validators.PathBuilder, selector map[string]string) validators.ValidationError {
				var err validators.ValidationError
				_, defined := selector[mesh_proto.ServiceTag]
				if len(selector) != 1 || !defined {
					err.AddViolationAt(path, fmt.Sprintf("must consist of exactly one tag %q", mesh_proto.ServiceTag))
				}
				return err
			},
		},
		ExtraTagKeyValidators: []TagKeyValidatorFunc{
			func(path validators.PathBuilder, key string) validators.ValidationError {
				var err validators.ValidationError
				if key != mesh_proto.ServiceTag {
					err.AddViolationAt(path.Key(key), fmt.Sprintf("tag %q is not allowed", key))
				}
				return err
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

func ValidateDuration(path validators.PathBuilder, duration *durationpb.Duration) validators.ValidationError {
	var errs validators.ValidationError
	if duration == nil {
		errs.AddViolationAt(path, "must have a positive value")
		return errs
	}
	if err := duration.CheckValid(); err != nil {
		errs.AddViolationAt(path, "must have a valid value")
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

// ValidateHostname validates a gateway hostname field. A hostname may be one of
//   - '*'
//   - '*.domain.name'
//   - 'domain.name'
func ValidateHostname(path validators.PathBuilder, hostname string) validators.ValidationError {
	if hostname == mesh_proto.WildcardHostname {
		return validators.ValidationError{}
	}

	err := validators.ValidationError{}

	if len(hostname) > 253 {
		err.AddViolationAt(path, "must be at most 253 characters")
	}

	if after, ok := strings.CutPrefix(hostname, "*."); ok {
		if !DomainRegexp.MatchString(after) {
			err.AddViolationAt(path, "invalid wildcard domain")
		}

		return err
	}

	if !DomainRegexp.MatchString(hostname) {
		err.AddViolationAt(path, "invalid hostname")
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

func ProtocolValidator(protocols core_meta.ProtocolList) TagsValidatorFunc {
	return func(path validators.PathBuilder, selector map[string]string) validators.ValidationError {
		var err validators.ValidationError
		v, defined := selector[mesh_proto.ProtocolTag]
		if !defined {
			err.AddViolationAt(path, "protocol must be specified")
			return err
		}

		if !protocols.Contains(core_meta.ParseProtocol(v)) {
			err.AddViolationAt(
				path.Key(mesh_proto.ProtocolTag),
				fmt.Sprintf("must be one of the [%s]", strings.Join(protocols.Strings(), ", ")),
			)
		}

		return err
	}
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

// SelectorKeyNotInSet returns a TagKeyValidatorFunc that checks the tag key
// is not any one of the given names.
func SelectorKeyNotInSet(keyName ...string) TagKeyValidatorFunc {
	set := map[string]struct{}{}

	for _, k := range keyName {
		set[k] = struct{}{}
	}

	return TagKeyValidatorFunc(
		func(path validators.PathBuilder, key string) validators.ValidationError {
			err := validators.ValidationError{}

			if _, ok := set[key]; ok {
				err.AddViolationAt(
					path.Key(key),
					fmt.Sprintf("tag name must not be %q", key),
				)
			}

			return err
		})
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
		err.Add(disallowedField("tags", pointer.Deref(ref.Tags), ref.Kind))
		err.Add(disallowedField("labels", pointer.Deref(ref.Labels), ref.Kind))
		err.Add(disallowedField("sectionName", pointer.Deref(ref.SectionName), ref.Kind))
	case common_api.Dataplane:
		err.Add(disallowedField("tags", pointer.Deref(ref.Tags), ref.Kind))
		if !opts.IsInboundPolicy && pointer.Deref(ref.SectionName) != "" {
			err.AddViolation("sectionName", "can only be used with inbound policies")
		}
	case common_api.LegacyMeshSubsetKind():
		err.Add(ValidateTags(validators.RootedAt("tags"), pointer.Deref(ref.Tags), ValidateTagsOpts{}))
		err.Add(disallowedField("labels", pointer.Deref(ref.Labels), ref.Kind))
		err.Add(disallowedField("sectionName", pointer.Deref(ref.SectionName), ref.Kind))
	case common_api.MeshService:
		err.Add(disallowedField("tags", pointer.Deref(ref.Tags), ref.Kind))
		err.Add(requiredField("labels", pointer.Deref(ref.Labels), ref.Kind))
	case common_api.MeshHTTPRoute:
		err.Add(disallowedField("tags", pointer.Deref(ref.Tags), ref.Kind))
		err.Add(disallowedField("sectionName", pointer.Deref(ref.SectionName), ref.Kind))
		err.Add(requiredField("labels", pointer.Deref(ref.Labels), ref.Kind))
	case common_api.LegacyMeshServiceSubsetKind():
		err.Add(requiredField("labels", pointer.Deref(ref.Labels), ref.Kind))
		err.Add(ValidateSelector(validators.RootedAt("tags"), pointer.Deref(ref.Tags), ValidateTagsOpts{}))
		err.Add(disallowedField("sectionName", pointer.Deref(ref.SectionName), ref.Kind))
	case common_api.MeshExternalService:
		err.Add(disallowedField("tags", pointer.Deref(ref.Tags), ref.Kind))
		err.Add(disallowedField("sectionName", pointer.Deref(ref.SectionName), ref.Kind))
		err.Add(requiredField("labels", pointer.Deref(ref.Labels), ref.Kind))
	case common_api.MeshMultiZoneService:
		err.Add(disallowedField("tags", pointer.Deref(ref.Tags), ref.Kind))
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
