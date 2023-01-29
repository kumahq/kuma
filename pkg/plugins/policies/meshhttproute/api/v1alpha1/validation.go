package v1alpha1

import (
	"fmt"
	"strings"

	common_api "github.com/kumahq/kuma/api/common/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/validators"
	matcher_validators "github.com/kumahq/kuma/pkg/plugins/policies/matchers/validators"
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
		errs.AddErrorAt(path.Field("filters"), validateFilters(rule.Default.Filters))
	}

	return errs
}

func validateMatches(matches []Match) validators.ValidationError {
	var errs validators.ValidationError

	for i, match := range matches {
		path := validators.Root().Index(i)
		errs.AddErrorAt(path.Field("path"), validatePath(match.Path))
		errs.AddErrorAt(path.Field("method"), validateMethod(match.Method))
		errs.AddErrorAt(path.Field("queryParams"), validateQueryParams(match.QueryParams))
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
	case Prefix:
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

func validateMethod(match *Method) validators.ValidationError {
	return validators.ValidationError{}
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

func validateFilters(filters *[]Filter) validators.ValidationError {
	var errs validators.ValidationError

	if filters == nil {
		return errs
	}

	for i, filter := range *filters {
		path := validators.Root().Index(i)
		switch filter.Type {
		case RequestHeaderModifierType:
			if filter.RequestHeaderModifier == nil {
				errs.AddViolationAt(path.Field("requestHeaderModifier"), validators.MustBeDefined)
			}
		case ResponseHeaderModifierType:
			if filter.ResponseHeaderModifier == nil {
				errs.AddViolationAt(path.Field("responseHeaderModifier"), validators.MustBeDefined)
			}
		case RequestRedirectType:
			if filter.RequestRedirect == nil {
				errs.AddViolationAt(path.Field("requestRedirect"), validators.MustBeDefined)
			}
		case URLRewriteType:
			if filter.URLRewrite == nil {
				errs.AddViolationAt(path.Field("urlRewrite"), validators.MustBeDefined)
			}
		}
	}

	return errs
}
