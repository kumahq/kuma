package mesh

import (
	"fmt"
	"strings"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/validators"
)

// Validate checks MeshGatewayRouteResource semantic constraints.
func (g *MeshGatewayRouteResource) Validate() error {
	var err validators.ValidationError

	err.Add(ValidateSelectors(
		validators.RootedAt("selectors"),
		g.Spec.GetSelectors(),
		ValidateSelectorsOpts{
			RequireAtLeastOneSelector: true,
			ValidateTagsOpts: ValidateTagsOpts{
				RequireAtLeastOneTag: true,
				RequireService:       true,
			},
		},
	))

	err.Add(validateMeshGatewayRouteConf(
		validators.RootedAt("conf"),
		g.Spec.GetConf(),
	))

	return err.OrNil()
}

func validateMeshGatewayRouteConf(path validators.PathBuilder, conf *mesh_proto.MeshGatewayRoute_Conf) validators.ValidationError {
	var err validators.ValidationError

	if conf.GetRoute() == nil {
		err.AddViolationAt(path, "cannot be empty")
	}

	err.Add(validateMeshGatewayRouteTCP(path.Field("tcp"), conf.GetTcp()))
	err.Add(validateMeshGatewayRouteHTTP(path.Field("http"), conf.GetHttp()))

	return err
}

func validateMeshGatewayRouteTCP(
	_ validators.PathBuilder,
	_ *mesh_proto.MeshGatewayRoute_TcpRoute,
) validators.ValidationError {
	return validators.OK()
}

func validateMeshGatewayRouteHTTP(
	path validators.PathBuilder,
	conf *mesh_proto.MeshGatewayRoute_HttpRoute,
) validators.ValidationError {
	if conf == nil {
		return validators.OK()
	}

	if len(conf.GetRules()) < 1 {
		return validators.MakeRequiredFieldErr(path.Field("rules"))
	}

	var err validators.ValidationError

	for i, rule := range conf.GetRules() {
		err.Add(validateMeshGatewayRouteHTTPRule(path.Field("rules").Index(i), rule))
	}

	return err
}

func validateMeshGatewayRouteHTTPRule(
	path validators.PathBuilder,
	conf *mesh_proto.MeshGatewayRoute_HttpRoute_Rule,
) validators.ValidationError {
	var hasRedirect bool
	var hasPrefixMatch bool

	if len(conf.GetMatches()) < 1 {
		return validators.MakeRequiredFieldErr(path.Field("matches"))
	}

	var err validators.ValidationError

	for i, m := range conf.GetMatches() {
		err.Add(validateMeshGatewayRouteHTTPMatch(path.Field("matches").Index(i), m))

		if p := m.GetPath(); p != nil && p.GetMatch() == mesh_proto.MeshGatewayRoute_HttpRoute_Match_Path_PREFIX {
			hasPrefixMatch = true
		}
	}

	var rewriteHostToBackendHostname bool
	var filters []*mesh_proto.MeshGatewayRoute_HttpRoute_Filter

	for _, f := range conf.GetFilters() {
		if m := f.GetRewrite(); m != nil {
			if m.GetHostToBackendHostname() {
				rewriteHostToBackendHostname = true
				continue
			}
		}

		filters = append(filters, f)
	}

	for i, f := range filters {
		if f.GetRedirect() != nil {
			hasRedirect = true
		}

		err.Add(validateMeshGatewayRouteHTTPFilter(path.Field("filters").Index(i), f, hasPrefixMatch, rewriteHostToBackendHostname))
	}

	// It doesn't make sense to redirect and also mirror or rewrite request headers.
	if hasRedirect && len(conf.GetFilters()) != 1 {
		err.AddViolationAt(path.Field("filters"), "redirects cannot be used with other filters")

		// Return since the redirect filter error makes the backend length check ambiguous.
		return err
	}

	switch len(conf.GetBackends()) {
	case 0:
		// Redirection doesn't forward, so there must not be any backend.
		if !hasRedirect {
			err.AddViolationAt(path.Field("backends"), "cannot be empty")
		}
	default:
		if hasRedirect {
			err.AddViolationAt(path.Field("backends"), "must be empty when using redirect filters")
		}
	}

	for i, b := range conf.GetBackends() {
		err.Add(validateMeshGatewayRouteBackend(path.Field("backends").Index(i), b))
	}

	return err
}

func validateMeshGatewayRouteHTTPMatch(
	path validators.PathBuilder,
	conf *mesh_proto.MeshGatewayRoute_HttpRoute_Match,
) validators.ValidationError {
	var err validators.ValidationError

	if conf.GetPath() == nil &&
		conf.GetMethod() == mesh_proto.HttpMethod_NONE &&
		len(conf.GetHeaders()) < 1 &&
		len(conf.GetQueryParameters()) < 1 {
		err.AddViolationAt(path, "cannot be empty")
	}

	if p := conf.GetPath(); p != nil {
		switch p.GetMatch() {
		case mesh_proto.MeshGatewayRoute_HttpRoute_Match_Path_REGEX:
			if p.GetValue() == "" {
				err.AddViolationAt(path.Field("value"), "cannot be empty")
			}
		case mesh_proto.MeshGatewayRoute_HttpRoute_Match_Path_PREFIX:
			if p.GetValue() == "/" {
				break
			}
			if !strings.HasPrefix(p.GetValue(), "/") {
				err.AddViolationAt(path.Field("value"), "must be an absolute path")
			}
		default:
			if !strings.HasPrefix(p.GetValue(), "/") {
				err.AddViolationAt(path.Field("value"), "must be an absolute path")
			}
		}
	}

	for i, h := range conf.GetHeaders() {
		path := path.Field("headers").Index(i)
		if h.GetName() == "" {
			err.AddViolationAt(path.Field("name"), "cannot be empty")
		}
		switch h.Match {
		case mesh_proto.MeshGatewayRoute_HttpRoute_Match_Header_ABSENT, mesh_proto.MeshGatewayRoute_HttpRoute_Match_Header_PRESENT:
			if h.GetValue() != "" {
				err.AddViolationAt(path.Field("value"), "cannot be set")
			}
		default:
			if h.GetValue() == "" {
				err.AddViolationAt(path.Field("value"), "cannot be empty")
			}
		}
	}

	for i, q := range conf.GetQueryParameters() {
		path := path.Field("query_parameters").Index(i)
		if q.GetName() == "" {
			err.AddViolationAt(path.Field("name"), "cannot be empty")
		}
		if q.GetValue() == "" {
			err.AddViolationAt(path.Field("value"), "cannot be empty")
		}
	}

	return err
}

func validateMeshGatewayRouteHTTPFilter(
	path validators.PathBuilder,
	conf *mesh_proto.MeshGatewayRoute_HttpRoute_Filter,
	hasPrefixMatch, rewriteHostToBackendHostname bool,
) validators.ValidationError {
	var err validators.ValidationError

	if r := conf.GetRequestHeader(); r != nil {
		header := func(
			path validators.PathBuilder,
			headers []*mesh_proto.MeshGatewayRoute_HttpRoute_Filter_HeaderFilter_Header,
		) validators.ValidationError {
			var err validators.ValidationError
			for i, h := range headers {
				switch h.GetName() {
				case ":authority", "Host", "host":
					if rewriteHostToBackendHostname {
						message := fmt.Sprintf(
							"cannot modify '%s' header, when route has set 'rewrite.host_to_backend_hostname' option",
							h.GetName(),
						)

						err.AddViolationAt(path.Index(i), message)
					}
				case "":
					err.AddViolationAt(path.Index(i).Field("name"), "cannot be empty")
				}
				if h.GetValue() == "" {
					err.AddViolationAt(path.Index(i).Field("value"), "cannot be empty")
				}
			}

			return err
		}

		path := path.Field("request_header")
		if len(r.GetSet()) < 1 &&
			len(r.GetAdd()) < 1 &&
			len(r.GetRemove()) < 1 {
			err.AddViolationAt(path, "cannot be empty")
		}

		err.Add(header(path.Field("set"), r.GetSet()))
		err.Add(header(path.Field("add"), r.GetAdd()))

		for i, h := range r.GetRemove() {
			if h == "" {
				err.AddViolationAt(path.Field("remove").Index(i), "cannot be empty")
			}
		}
	}

	if r := conf.GetRedirect(); r != nil {
		path := path.Field("redirect")

		if r.GetPort() > 0 {
			err.Add(ValidatePort(path.Field("port"), r.GetPort()))
		}

		if r.GetStatusCode() < 300 || r.GetStatusCode() > 308 {
			err.AddViolationAt(path.Field("status_code"), "must be in the range [300, 308]")
		}
	}

	if m := conf.GetMirror(); m != nil {
		path := path.Field("mirror")

		err.Add(validatePercentage(path, m.GetPercentage()))
		err.Add(validateMeshGatewayRouteBackend(
			path.Field("backend"),
			m.GetBackend(),
		))
	}

	if m := conf.GetRewrite(); m != nil {
		path := path.Field("rewrite")
		if _, ok := m.GetPath().(*mesh_proto.MeshGatewayRoute_HttpRoute_Filter_Rewrite_ReplacePrefixMatch); ok && !hasPrefixMatch {
			err.AddViolationAt(path.Field("replacePrefixMatch"), "cannot be used without a match on path prefix")
		}
	}

	return err
}

func validateMeshGatewayRouteBackend(
	path validators.PathBuilder,
	conf *mesh_proto.MeshGatewayRoute_Backend,
) validators.ValidationError {
	if conf == nil {
		return validators.MakeRequiredFieldErr(path)
	}

	return ValidateSelector(
		path,
		conf.GetDestination(),
		ValidateTagsOpts{
			RequireAtLeastOneTag: true,
			RequireService:       true,
		})
}
