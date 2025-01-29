package mesh

import (
	"fmt"
	"regexp"
	"slices"
	"sort"
	"strings"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/durationpb"
	"sigs.k8s.io/yaml"

	common_api "github.com/kumahq/kuma/api/common/v1alpha1"
	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/validators"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
)

const dnsLabel = `[a-z0-9]([-a-z0-9]*[a-z0-9])?`

var (
	NameCharacterSet     = regexp.MustCompile(`^[0-9a-z.\-_]*$`)
	tagNameCharacterSet  = regexp.MustCompile(`^[a-zA-Z0-9.\-_:/]*$`)
	tagValueCharacterSet = regexp.MustCompile(`^[a-zA-Z0-9.\-_:]*$`)
	selectorCharacterSet = regexp.MustCompile(`^([a-zA-Z0-9.\-_:/]*|\*)$`)
	domainRegexp         = regexp.MustCompile("^" + dnsLabel + "(\\." + dnsLabel + ")*" + "$")
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
	// AllowedInvalidNames field allows to provide names that deviate from
	// standard naming conventions in specific scenarios. I.e. normally,
	// service names cannot contain forward slashes ("/"). However, there
	// are exceptions during resource conversion:
	// * Gateway API to Kuma HTTPRoute Conversion
	//   When converting an HTTPRoute from Gateway API to a MeshHTTPRoute
	//   (Kuma's resource definition), there might be situations where the
	//   targeted backend reference cannot be found. In such cases, Kuma
	//   sets the service name to "kuma.io/unresolved-backend". This name
	//   includes a forward slash, but it's allowed as an exception to
	//   handle unresolved references.
	AllowedInvalidNames []string
	IsInboundPolicy     bool
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

	if strings.HasPrefix(hostname, "*.") {
		if !domainRegexp.MatchString(strings.TrimPrefix(hostname, "*.")) {
			err.AddViolationAt(path, "invalid wildcard domain")
		}

		return err
	}

	if !domainRegexp.MatchString(hostname) {
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

func ProtocolValidator(protocols ...string) TagsValidatorFunc {
	return func(path validators.PathBuilder, selector map[string]string) validators.ValidationError {
		var err validators.ValidationError
		v, defined := selector[mesh_proto.ProtocolTag]
		if !defined {
			err.AddViolationAt(path, "protocol must be specified")
			return err
		}
		for _, protocol := range protocols {
			if v == protocol {
				return err
			}
		}
		err.AddViolationAt(path.Key(mesh_proto.ProtocolTag), fmt.Sprintf("must be one of the [%s]",
			strings.Join(protocols, ", ")))
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
		errMsg := "value is not supported"
		if optsErr := opts.SupportedKindsError; optsErr != "" {
			errMsg = optsErr
		}
		err.AddViolation("kind", errMsg)
		return err
	}

	switch ref.Kind {
	case common_api.Mesh:
		if ref.Name != "" {
			err.AddViolation("name", fmt.Sprintf("using name with kind %v is not yet supported", ref.Kind))
		}
		err.Add(disallowedField("mesh", ref.Mesh, ref.Kind))
		err.Add(disallowedField("tags", ref.Tags, ref.Kind))
		err.Add(disallowedField("labels", ref.Labels, ref.Kind))
		err.Add(disallowedField("namespace", ref.Namespace, ref.Kind))
		err.Add(disallowedField("sectionName", ref.SectionName, ref.Kind))
	case common_api.Dataplane:
		err.Add(disallowedField("tags", ref.Tags, ref.Kind))
		err.Add(disallowedField("mesh", ref.Mesh, ref.Kind))
		err.Add(disallowedField("proxyTypes", ref.ProxyTypes, ref.Kind))
		if len(ref.Labels) > 0 && (ref.Name != "" || ref.Namespace != "") {
			err.AddViolation("labels", "either labels or name and namespace must be specified")
		}
		if !opts.IsInboundPolicy && ref.SectionName != "" {
			err.AddViolation("sectionName", "can only be used with inbound policies")
		}
	case common_api.MeshSubset:
		err.Add(disallowedField("name", ref.Name, ref.Kind))
		err.Add(disallowedField("mesh", ref.Mesh, ref.Kind))
		err.Add(ValidateTags(validators.RootedAt("tags"), ref.Tags, ValidateTagsOpts{}))
		err.Add(disallowedField("labels", ref.Labels, ref.Kind))
		err.Add(disallowedField("namespace", ref.Namespace, ref.Kind))
		err.Add(disallowedField("sectionName", ref.SectionName, ref.Kind))
	case common_api.MeshService, common_api.MeshHTTPRoute:
		err.Add(validateName(ref.Name, opts.AllowedInvalidNames))
		err.Add(disallowedField("mesh", ref.Mesh, ref.Kind))
		err.Add(disallowedField("tags", ref.Tags, ref.Kind))
		err.Add(disallowedField("proxyTypes", ref.ProxyTypes, ref.Kind))
		if len(ref.Labels) == 0 {
			err.Add(requiredField("name", ref.Name, ref.Kind))
		}
		if len(ref.Labels) > 0 && (ref.Name != "" || ref.Namespace != "") {
			err.AddViolation("labels", "either labels or name and namespace must be specified")
		}
		if len(ref.Labels) > 0 && ref.SectionName != "" {
			err.AddViolation("sectionName", "sectionName should not be combined with labels")
		}
	case common_api.MeshServiceSubset, common_api.MeshGateway:
		err.Add(requiredField("name", ref.Name, ref.Kind))
		err.Add(validateName(ref.Name, opts.AllowedInvalidNames))
		err.Add(disallowedField("mesh", ref.Mesh, ref.Kind))
		err.Add(disallowedField("proxyTypes", ref.ProxyTypes, ref.Kind))
		err.Add(ValidateSelector(validators.RootedAt("tags"), ref.Tags, ValidateTagsOpts{}))
		if ref.Kind == common_api.MeshGateway && len(ref.Tags) > 0 && !opts.GatewayListenerTagsAllowed {
			err.Add(disallowedField("tags", ref.Tags, ref.Kind))
		}
		err.Add(disallowedField("labels", ref.Labels, ref.Kind))
		err.Add(disallowedField("namespace", ref.Namespace, ref.Kind))
		err.Add(disallowedField("sectionName", ref.SectionName, ref.Kind))
	case common_api.MeshExternalService:
		err.Add(validateName(ref.Name, opts.AllowedInvalidNames))
		err.Add(disallowedField("mesh", ref.Mesh, ref.Kind))
		err.Add(disallowedField("tags", ref.Tags, ref.Kind))
		err.Add(disallowedField("proxyTypes", ref.ProxyTypes, ref.Kind))
		if len(ref.Labels) == 0 {
			err.Add(requiredField("name", ref.Name, ref.Kind))
		}
		if len(ref.Labels) > 0 && (ref.Name != "" || ref.Namespace != "") {
			err.AddViolation("labels", "either labels or name must be specified")
		}
		err.Add(disallowedField("sectionName", ref.SectionName, ref.Kind))
	}

	return err
}

func validateName(value string, allowedInvalidNames []string) validators.ValidationError {
	var err validators.ValidationError

	if !slices.Contains(allowedInvalidNames, value) && !NameCharacterSet.MatchString(value) {
		err.AddViolation(
			"name",
			"invalid characters: must consist of lower case alphanumeric characters, '-', '.' and '_'.",
		)
	}

	return err
}

func disallowedField[T ~string | ~map[string]string | ~[]common_api.TargetRefProxyType](
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
		err.AddViolation(name, fmt.Sprintf("%s with kind %v", validators.MustBeSet, kind))
	}

	return err
}

func isSet[T ~string | ~map[string]string | ~[]common_api.TargetRefProxyType](value T) bool {
	switch v := any(value).(type) {
	case string:
		return v != ""
	case map[string]string:
		return len(v) > 0
	case []common_api.TargetRefProxyType:
		return len(v) > 0
	default:
		return false
	}
}
