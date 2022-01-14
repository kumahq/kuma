package mesh

import (
	"bytes"
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/ghodss/yaml"
	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
	"google.golang.org/protobuf/types/known/durationpb"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/validators"
)

type TagsValidatorFunc func(path validators.PathBuilder, selector map[string]string) validators.ValidationError
type TagKeyValidatorFunc func(path validators.PathBuilder, key string) validators.ValidationError
type TagValueValidatorFunc func(path validators.PathBuilder, key, value string) validators.ValidationError

type ValidateTagsOpts struct {
	RequireAtLeastOneTag    bool
	RequireService          bool
	ExtraTagsValidators     []TagsValidatorFunc
	ExtraTagKeyValidators   []TagKeyValidatorFunc
	ExtraTagValueValidators []TagValueValidatorFunc
}

type ValidateSelectorsOpts struct {
	ValidateTagsOpts
	RequireAtLeastOneSelector bool
}

var tagNameCharacterSet = regexp.MustCompile(`^[a-zA-Z0-9\.\-_:/]*$`)
var tagValueCharacterSet = regexp.MustCompile(`^[a-zA-Z0-9\.\-_:]*$`)
var selectorCharacterSet = regexp.MustCompile(`^([a-zA-Z0-9\.\-_:/]*|\*)$`)

func ValidateSelectors(path validators.PathBuilder, sources []*mesh_proto.Selector, opts ValidateSelectorsOpts) (err validators.ValidationError) {
	if opts.RequireAtLeastOneSelector && len(sources) == 0 {
		err.AddViolationAt(path, "must have at least one element")
	}
	for i, selector := range sources {
		err.Add(ValidateSelector(path.Index(i).Field("match"), selector.GetMatch(), opts.ValidateTagsOpts))
	}
	return
}

func ValidateSelector(path validators.PathBuilder, tags map[string]string, opts ValidateTagsOpts) (err validators.ValidationError) {
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

func ValidateTags(path validators.PathBuilder, tags map[string]string, opts ValidateTagsOpts) (err validators.ValidationError) {
	opts.ExtraTagValueValidators = append([]TagValueValidatorFunc{
		func(path validators.PathBuilder, key, value string) validators.ValidationError {
			var err validators.ValidationError
			if !tagValueCharacterSet.MatchString(value) {
				err.AddViolationAt(path.Key(key), `tag value must consist of alphanumeric characters, dots, dashes and underscores`)
			}
			return err
		},
	}, opts.ExtraTagValueValidators...)

	return validateTagKeyValues(path, tags, opts)
}

func validateTagKeyValues(path validators.PathBuilder, keyValues map[string]string, opts ValidateTagsOpts) (err validators.ValidationError) {
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
			err.AddViolationAt(path.Key(key), `tag name must consist of alphanumeric characters, dots, dashes, slashes and underscores`)
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
	if opts.RequireService && !defined {
		err.AddViolationAt(path, fmt.Sprintf("mandatory tag %q is missing", mesh_proto.ServiceTag))
	}
	return
}

var OnlyServiceTagAllowed = ValidateSelectorsOpts{
	RequireAtLeastOneSelector: true,
	ValidateTagsOpts: ValidateTagsOpts{
		RequireService: true,
		ExtraTagsValidators: []TagsValidatorFunc{
			func(path validators.PathBuilder, selector map[string]string) (err validators.ValidationError) {
				_, defined := selector[mesh_proto.ServiceTag]
				if len(selector) != 1 || !defined {
					err.AddViolationAt(path, fmt.Sprintf("must consist of exactly one tag %q", mesh_proto.ServiceTag))
				}
				return
			},
		},
		ExtraTagKeyValidators: []TagKeyValidatorFunc{
			func(path validators.PathBuilder, key string) (err validators.ValidationError) {
				if key != mesh_proto.ServiceTag {
					err.AddViolationAt(path.Key(key), fmt.Sprintf("tag %q is not allowed", key))
				}
				return
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

func ValidateDuration(path validators.PathBuilder, duration *durationpb.Duration) (errs validators.ValidationError) {
	if duration == nil {
		errs.AddViolationAt(path, "must have a positive value")
		return
	}
	if err := duration.CheckValid(); err != nil {
		errs.AddViolationAt(path, "must have a valid value")
		return
	}
	if duration.AsDuration() == 0 {
		errs.AddViolationAt(path, "must have a positive value")
	}
	return
}

func ValidateThreshold(path validators.PathBuilder, threshold uint32) (err validators.ValidationError) {
	if threshold == 0 {
		err.AddViolationAt(path, "must have a positive value")
	}
	return
}

// ValidatePort validates that port is a valid TCP or UDP port number.
func ValidatePort(path validators.PathBuilder, port uint32) validators.ValidationError {
	err := validators.ValidationError{}

	if port == 0 || port > 65535 {
		err.AddViolationAt(path, "port must be in the range [1, 65535]")
	}

	return err
}

const dnsLabel = `[a-z0-9]([-a-z0-9]*[a-z0-9])?`

var domainRegexp = regexp.MustCompile("^" + dnsLabel + "(\\." + dnsLabel + ")*" + "$")

// ValidateHostname validates a gateway hostname field. A hostname may be one of
//	- '*'
//	- '*.domain.name'
//	- 'domain.name'
func ValidateHostname(path validators.PathBuilder, hostname string) validators.ValidationError {
	if hostname == "*" {
		return validators.ValidationError{}
	}

	err := validators.ValidationError{}

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
	return func(path validators.PathBuilder, selector map[string]string) (err validators.ValidationError) {
		v, defined := selector[mesh_proto.ProtocolTag]
		if !defined {
			err.AddViolationAt(path, "protocol must be specified")
			return
		}
		for _, protocol := range protocols {
			if v == protocol {
				return
			}
		}
		err.AddViolationAt(path.Key(mesh_proto.ProtocolTag), fmt.Sprintf("must be one of the [%s]",
			strings.Join(protocols, ", ")))
		return
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

	if err := (&jsonpb.Unmarshaler{}).Unmarshal(bytes.NewReader(json), msg); err != nil {
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
	return (&jsonpb.Unmarshaler{}).Unmarshal(bytes.NewReader(json), msg)
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
