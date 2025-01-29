package v1alpha1

import (
	"fmt"
	"slices"
	"strings"

	k8s_validation "k8s.io/apimachinery/pkg/util/validation"
	netutils "k8s.io/utils/net"

	common_api "github.com/kumahq/kuma/api/common/v1alpha1"
	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/validators"
	"github.com/kumahq/kuma/pkg/plugins/runtime/gateway/metadata"
	"github.com/kumahq/kuma/pkg/util/pointer"
)

func (r *MeshHTTPRouteResource) validate() error {
	var verr validators.ValidationError
	path := validators.RootedAt("spec")
	verr.AddErrorAt(path.Field("targetRef"), r.validateTop(r.Spec.TargetRef))
	verr.AddErrorAt(path.Field("to"), validateTos(pointer.DerefOr(r.Spec.TargetRef, common_api.TargetRef{Kind: common_api.Mesh}), r.Spec.To))
	return verr.OrNil()
}

func (r *MeshHTTPRouteResource) validateTop(targetRef *common_api.TargetRef) validators.ValidationError {
	if targetRef == nil {
		return validators.ValidationError{}
	}
	switch core_model.PolicyRole(r.GetMeta()) {
	case mesh_proto.SystemPolicyRole:
		return mesh.ValidateTargetRef(*targetRef, &mesh.ValidateTargetRefOpts{
			SupportedKinds: []common_api.TargetRefKind{
				common_api.Mesh,
				common_api.MeshGateway,
				common_api.MeshSubset,
				common_api.MeshService,
				common_api.MeshServiceSubset,
				common_api.Dataplane,
			},
			GatewayListenerTagsAllowed: true,
		})
	default:
		return mesh.ValidateTargetRef(*targetRef, &mesh.ValidateTargetRefOpts{
			SupportedKinds: []common_api.TargetRefKind{
				common_api.Mesh,
				common_api.MeshSubset,
				common_api.MeshService,
				common_api.MeshServiceSubset,
				common_api.Dataplane,
			},
		})
	}
}

func validateToRef(topTargetRef, targetRef common_api.TargetRef) validators.ValidationError {
	switch topTargetRef.Kind {
	case common_api.MeshGateway:
		return mesh.ValidateTargetRef(targetRef, &mesh.ValidateTargetRefOpts{
			SupportedKinds: []common_api.TargetRefKind{
				common_api.Mesh,
			},
		})
	default:
		return mesh.ValidateTargetRef(targetRef, &mesh.ValidateTargetRefOpts{
			SupportedKinds: []common_api.TargetRefKind{
				common_api.MeshService,
				common_api.MeshExternalService,
				common_api.MeshMultiZoneService,
			},
		})
	}
}

func validateTos(topTargetRef common_api.TargetRef, tos []To) validators.ValidationError {
	var errs validators.ValidationError

	for i, to := range tos {
		path := validators.Root().Index(i)
		errs.AddErrorAt(path.Field("targetRef"), validateToRef(topTargetRef, to.TargetRef))
		errs.AddErrorAt(path.Field("rules"), validateRules(topTargetRef, to.Rules))
		errs.AddErrorAt(path.Field("hostnames"), validateHostnames(topTargetRef, to.Hostnames))
	}

	return errs
}

func validateRules(topTargetRef common_api.TargetRef, rules []Rule) validators.ValidationError {
	var errs validators.ValidationError

	for i, rule := range rules {
		path := validators.Root().Index(i)
		errs.AddErrorAt(path.Field("matches"), validateMatches(rule.Matches))
		errs.AddErrorAt(path.Field("default").Field("filters"), validateFilters(topTargetRef, rule.Default.Filters, rule.Matches))
		errs.AddErrorAt(path.Field("default").Field("backendRefs"), validateBackendRefs(
			topTargetRef,
			pointer.Deref(rule.Default.BackendRefs),
			pointer.Deref(rule.Default.Filters),
		))
	}

	return errs
}

func validateHostnames(topTargetRef common_api.TargetRef, hostnames []string) validators.ValidationError {
	var errs validators.ValidationError

	path := validators.Root()
	switch topTargetRef.Kind {
	case common_api.MeshGateway:
	default:
		if len(hostnames) > 0 {
			errs.AddViolationAt(path, validators.MustNotBeDefined)
		}
	}
	return errs
}

func validateMatches(matches []Match) validators.ValidationError {
	var errs validators.ValidationError

	for i, match := range matches {
		path := validators.Root().Index(i)
		errs.AddErrorAt(path.Field("path"), validatePath(match.Path))
		errs.AddErrorAt(path.Field("queryParams"), validateQueryParams(match.QueryParams))
		errs.AddErrorAt(path.Field("headers"), validateHeaders(match.Headers))
	}

	return errs
}

func validatePath(match *PathMatch) validators.ValidationError {
	var errs validators.ValidationError

	if match == nil {
		return errs
	}

	valuePath := validators.RootedAt("value")

	switch match.Type {
	case RegularExpression:
		break
	case PathPrefix:
		if match.Value == "/" {
			break
		}
		if strings.HasSuffix(match.Value, "/") {
			errs.AddViolationAt(valuePath, "does not need a trailing slash because only a `/`-separated prefix or an entire path is matched")
		}
		if !strings.HasPrefix(match.Value, "/") {
			errs.AddViolationAt(valuePath, "must be an absolute path")
		}
	case Exact:
		if !strings.HasPrefix(match.Value, "/") {
			errs.AddViolationAt(valuePath, "must be an absolute path")
		}
	}

	return errs
}

func validateHeaders(headers []common_api.HeaderMatch) validators.ValidationError {
	var errs validators.ValidationError
	for i, header := range headers {
		path := validators.Root().Index(i)
		matchType := common_api.HeaderMatchExact
		if header.Type != nil {
			matchType = *header.Type
		}

		switch matchType {
		case common_api.HeaderMatchExact:
		case common_api.HeaderMatchPresent:
			if header.Value != "" {
				errs.AddViolationAt(path.Field("value"), validators.MustNotBeDefined)
			}
		case common_api.HeaderMatchRegularExpression:
		case common_api.HeaderMatchAbsent:
			if header.Value != "" {
				errs.AddViolationAt(path.Field("value"), validators.MustNotBeDefined)
			}
		case common_api.HeaderMatchPrefix:
		}
	}
	return errs
}

func validateQueryParams(matches []QueryParamsMatch) validators.ValidationError {
	var errs validators.ValidationError

	matchedNames := map[string]struct{}{}
	for i, match := range matches {
		if _, ok := matchedNames[match.Name]; ok {
			path := validators.Root().Index(i).Field("name")
			errs.AddViolationAt(path, fmt.Sprintf("multiple entries for name %s", match.Name))
		}
		matchedNames[match.Name] = struct{}{}
	}

	return errs
}

func hasAnyMatchesWithoutPrefix(matches []Match) bool {
	for _, match := range matches {
		if match.Path == nil || match.Path.Type != PathPrefix {
			return true
		}
	}

	// No matches means a default / prefix
	return false
}

func validateFilters(topTargetRef common_api.TargetRef, filters *[]Filter, matches []Match) validators.ValidationError {
	var errs validators.ValidationError

	if filters == nil {
		return errs
	}

	for i, filter := range *filters {
		path := validators.Root().Index(i)
		switch filter.Type {
		case RequestHeaderModifierType:
			field := path.Field("requestHeaderModifier")
			if filter.RequestHeaderModifier == nil {
				errs.AddViolationAt(field, validators.MustBeDefined)
			} else {
				errs.AddErrorAt(field, validateHeaderModifier(filter.RequestHeaderModifier))
			}
		case ResponseHeaderModifierType:
			field := path.Field("responseHeaderModifier")
			if filter.ResponseHeaderModifier == nil {
				errs.AddViolationAt(field, validators.MustBeDefined)
			} else {
				errs.AddErrorAt(field, validateHeaderModifier(filter.ResponseHeaderModifier))
			}
		case RequestRedirectType:
			if filter.RequestRedirect == nil {
				errs.AddViolationAt(path.Field("requestRedirect"), validators.MustBeDefined)
				continue
			}
			if filter.RequestRedirect.Path != nil &&
				filter.RequestRedirect.Path.ReplacePrefixMatch != nil &&
				hasAnyMatchesWithoutPrefix(matches) {
				errs.AddViolationAt(path.Field("requestRedirect").Field("path").Field("replacePrefixMatch"), "can only appear if all matches match a path prefix")
			}
			errs.AddErrorAt(path.Field("requestRedirect").Field("hostname"), validatePreciseHostname(filter.RequestRedirect.Hostname))
		case URLRewriteType:
			if filter.URLRewrite == nil {
				errs.AddViolationAt(path.Field("urlRewrite"), validators.MustBeDefined)
				continue
			}
			if filter.URLRewrite.Path != nil &&
				filter.URLRewrite.Path.ReplacePrefixMatch != nil &&
				hasAnyMatchesWithoutPrefix(matches) {
				errs.AddViolationAt(path.Field("urlRewrite").Field("path").Field("replacePrefixMatch"), "can only appear if all matches match a path prefix")
			}
			errs.AddErrorAt(path.Field("urlRewrite").Field("hostname"), validatePreciseHostname(filter.URLRewrite.Hostname))
			if filter.URLRewrite.HostToBackendHostname {
				if topTargetRef.Kind != common_api.MeshGateway {
					errs.AddViolationAt(path.Field("urlRewrite").Field("hostToBackendHostname"), "can only be set with MeshGateway")
				}

				if filter.URLRewrite.Hostname != nil {
					errs.AddViolationAt(path.Field("urlRewrite").Field("hostToBackendHostname"), "cannot be set together with hostname")
				}
			}
		case RequestMirrorType:
			if filter.RequestMirror == nil {
				errs.AddViolationAt(path.Field("requestMirror"), validators.MustBeDefined)
				continue
			}
			errs.AddErrorAt(
				path.Field("requestMirror").Field("backendRef"),
				mesh.ValidateTargetRef(filter.RequestMirror.BackendRef.TargetRef, &mesh.ValidateTargetRefOpts{
					SupportedKinds: []common_api.TargetRefKind{
						common_api.MeshService,
						common_api.MeshServiceSubset,
					},
				}),
			)
		}
	}

	return errs
}

// PreciseHostname is the fully qualified domain name of a network host. This
// matches the RFC 1123 definition of a hostname with 1 notable exception that
// numeric IP addresses are not allowed.
//
// Note that as per RFC1035 and RFC1123, a *label* must consist of lower case
// alphanumeric characters or '-', and must start and end with an alphanumeric
// character. No other punctuation is allowed.
func validatePreciseHostname(hostname *PreciseHostname) validators.ValidationError {
	var errs validators.ValidationError

	if hostname == nil {
		return errs
	}

	if netutils.ParseIPSloppy(string(*hostname)) != nil {
		errs.AddViolationAt(validators.Root(), "cannot be an IP address")
		return errs
	}

	for _, violation := range k8s_validation.IsDNS1123Subdomain(string(*hostname)) {
		errs.AddViolationAt(validators.Root(), violation)
	}

	return errs
}

func validateHeaderModifier(modifier *HeaderModifier) validators.ValidationError {
	var errs validators.ValidationError

	if modifier.Set == nil && modifier.Add == nil && modifier.Remove == nil {
		errs.AddViolationAt(validators.Root(), validators.MustHaveAtLeastOne("set", "add", "remove"))
		return errs
	}

	headerSet := map[common_api.HeaderName]struct{}{}

	add := func(header common_api.HeaderName) validators.ValidationError {
		var verrs validators.ValidationError
		if _, ok := headerSet[header]; ok {
			verrs.AddViolation("name", "duplicate header name")
		}
		headerSet[header] = struct{}{}
		return verrs
	}

	for i, hm := range modifier.Set {
		errs.AddErrorAt(validators.Root().Field("set").Index(i), add(hm.Name))
	}

	for i, hm := range modifier.Add {
		errs.AddErrorAt(validators.Root().Field("add").Index(i), add(hm.Name))
	}

	for i, name := range modifier.Remove {
		errs.AddErrorAt(validators.Root().Field("remove").Index(i), add(common_api.HeaderName(name)))
	}

	return errs
}

func validateBackendRefs(
	topTargetRef common_api.TargetRef,
	backendRefs []common_api.BackendRef,
	filters []Filter,
) validators.ValidationError {
	var errs validators.ValidationError

	if len(backendRefs) == 0 {
		if topTargetRef.Kind == common_api.MeshGateway {
			containsBackendlessFilter := slices.ContainsFunc(filters, func(filter Filter) bool {
				return filter.Type == RequestRedirectType
			})

			// Rule doesn't need to contain any backendRefs when it contains RequestRedirectType filter
			if !containsBackendlessFilter {
				errs.AddViolationAt(validators.Root(), validators.MustNotBeEmpty)
			}
		}
	}

	for i, backendRef := range backendRefs {
		errs.AddErrorAt(
			validators.Root().Index(i),
			mesh.ValidateTargetRef(backendRef.TargetRef, &mesh.ValidateTargetRefOpts{
				SupportedKinds: []common_api.TargetRefKind{
					common_api.MeshService,
					common_api.MeshServiceSubset,
					common_api.MeshExternalService,
					common_api.MeshMultiZoneService,
				},
				AllowedInvalidNames: []string{metadata.UnresolvedBackendServiceTag},
			}),
		)
		errs.AddErrorAt(
			validators.Root().Index(i),
			validators.ValidateBackendRef(backendRef),
		)
	}

	return errs
}
