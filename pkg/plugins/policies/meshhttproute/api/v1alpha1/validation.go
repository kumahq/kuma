package v1alpha1

import (
	"fmt"
	matcher_validators "github.com/kumahq/kuma/pkg/plugins/policies/core/matchers/validators"
	"strings"

	common_api "github.com/kumahq/kuma/api/common/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/validators"
)

func (r *MeshHTTPRouteResource) validate() error {
	var verr validators.ValidationError
	path := validators.RootedAt("spec")
	verr.AddErrorAt(path.Field("targetRef"), validateTop(r.Spec.TargetRef))
	verr.AddErrorAt(path.Field("to"), validateTos(r.Spec.To))
	return verr.OrNil()
}

func validateTop(targetRef common_api.TargetRef) validators.ValidationError {
	return matcher_validators.ValidateTargetRef(targetRef, &matcher_validators.ValidateTargetRefOpts{
		SupportedKinds: []common_api.TargetRefKind{
			common_api.Mesh,
			common_api.MeshSubset,
			common_api.MeshService,
			common_api.MeshServiceSubset,
		},
	})
}

func validateToRef(targetRef common_api.TargetRef) validators.ValidationError {
	return matcher_validators.ValidateTargetRef(targetRef, &matcher_validators.ValidateTargetRefOpts{
		SupportedKinds: []common_api.TargetRefKind{
			common_api.MeshService,
		},
	})
}

func validateTos(tos []To) validators.ValidationError {
	var errs validators.ValidationError

	for i, to := range tos {
		path := validators.Root().Index(i)
		errs.AddErrorAt(path.Field("targetRef"), validateToRef(to.TargetRef))
		errs.AddErrorAt(path.Field("rules"), validateRules(to.Rules))
	}

	return errs
}

func validateRules(rules []Rule) validators.ValidationError {
	var errs validators.ValidationError

	for i, rule := range rules {
		path := validators.Root().Index(i)
		errs.AddErrorAt(path.Field("matches"), validateMatches(rule.Matches))
		errs.AddErrorAt(path.Field("default").Field("filters"), validateFilters(rule.Default.Filters, rule.Matches))
		errs.AddErrorAt(path.Field("default").Field("backendRefs"), validateBackendRefs(rule.Default.BackendRefs))
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

func validateFilters(filters *[]Filter, matches []Match) validators.ValidationError {
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
		case RequestMirrorType:
			if filter.RequestMirror == nil {
				errs.AddViolationAt(path.Field("requestMirror"), validators.MustBeDefined)
				continue
			}
			errs.AddErrorAt(
				path.Field("requestMirror").Field("backendRef"),
				matcher_validators.ValidateTargetRef(filter.RequestMirror.BackendRef, &matcher_validators.ValidateTargetRefOpts{
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

func validateBackendRefs(backendRefs *[]common_api.BackendRef) validators.ValidationError {
	var errs validators.ValidationError

	if backendRefs == nil {
		return errs
	}

	for i, backendRef := range *backendRefs {
		errs.AddErrorAt(
			validators.Root().Index(i),
			matcher_validators.ValidateTargetRef(backendRef.TargetRef, &matcher_validators.ValidateTargetRefOpts{
				SupportedKinds: []common_api.TargetRefKind{
					common_api.MeshService,
					common_api.MeshServiceSubset,
				},
			}),
		)
	}

	return errs
}
