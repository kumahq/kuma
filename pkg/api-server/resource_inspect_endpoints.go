package api_server

import (
	"errors"
	"fmt"
	"slices"
	"sort"
	"strconv"
	"strings"

	"github.com/emicklei/go-restful/v3"

	common_api "github.com/kumahq/kuma/v3/api/common/v1alpha1"
	api_types "github.com/kumahq/kuma/v3/api/openapi/types"
	api_common "github.com/kumahq/kuma/v3/api/openapi/types/common"
	oapi_helpers "github.com/kumahq/kuma/v3/pkg/api-server/oapi-helpers"
	"github.com/kumahq/kuma/v3/pkg/core/kri"
	core_plugins "github.com/kumahq/kuma/v3/pkg/core/plugins"
	"github.com/kumahq/kuma/v3/pkg/core/resources/access"
	core_mesh "github.com/kumahq/kuma/v3/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/v3/pkg/core/resources/manager"
	core_model "github.com/kumahq/kuma/v3/pkg/core/resources/model"
	"github.com/kumahq/kuma/v3/pkg/core/resources/registry"
	"github.com/kumahq/kuma/v3/pkg/core/resources/store"
	rest_errors "github.com/kumahq/kuma/v3/pkg/core/rest/errors"
	"github.com/kumahq/kuma/v3/pkg/core/user"
	core_xds "github.com/kumahq/kuma/v3/pkg/core/xds"
	"github.com/kumahq/kuma/v3/pkg/core/xds/inspect"
	"github.com/kumahq/kuma/v3/pkg/plugins/policies/core/matchers"
	core_rules "github.com/kumahq/kuma/v3/pkg/plugins/policies/core/rules"
	"github.com/kumahq/kuma/v3/pkg/plugins/policies/core/rules/common"
	"github.com/kumahq/kuma/v3/pkg/plugins/policies/core/rules/outbound"
	meshhttproute_api "github.com/kumahq/kuma/v3/pkg/plugins/policies/meshhttproute/api/v1alpha1"
	meshtcproute_api "github.com/kumahq/kuma/v3/pkg/plugins/policies/meshtcproute/api/v1alpha1"
	"github.com/kumahq/kuma/v3/pkg/util/pointer"
	util_slices "github.com/kumahq/kuma/v3/pkg/util/slices"
	xds_context "github.com/kumahq/kuma/v3/pkg/xds/context"
	xds_hooks "github.com/kumahq/kuma/v3/pkg/xds/hooks"
)

// resourceInspectHandler serves the resource inspect endpoints (rules, policies,
// proxy config, matching dataplanes).
type resourceInspectHandler struct {
	resourceEndpointsContext

	meshContextBuilder     xds_context.MeshContextBuilder
	xdsHooks               []xds_hooks.ResourceSetHook
	knownInternalAddresses []string
}

func (r *resourceInspectHandler) matchingDataplanesForPolicy() handlerFunc {
	return func(request *restful.Request) (any, error) {
		meshName, err := r.meshFromRequest(request)
		if err != nil {
			return nil, withTitle(err, "Failed to retrieve Mesh")
		}

		var dependentTypes []core_model.ResourceType
		if r.descriptor.IsTargetRefBased {
			dependentTypes = []core_model.ResourceType{meshhttproute_api.MeshHTTPRouteType}
		}
		dependentResources := xds_context.NewResources()
		for _, dependentType := range dependentTypes {
			hl, err := registry.Global().NewList(dependentType)
			if err != nil {
				return nil, withTitle(err, "failed inspect")
			}
			if err := r.resManager.List(request.Request.Context(), hl, store.ListByMesh(meshName)); err != nil {
				return nil, withTitle(err, "failed inspect")
			}
			dependentResources.MeshLocalResources[dependentType] = hl
		}
		return matchingDataplanesForFilter(
			request,
			r.descriptor,
			r.resManager,
			r.resourceAccess,
			func(policyResource core_model.Resource) store.ListFilterFunc {
				return func(rs core_model.Resource) bool {
					dpp := rs.(*core_mesh.DataplaneResource)
					res, _ := matchers.PolicyMatches(policyResource, dpp, dependentResources)
					return res
				}
			},
		)
	}
}

func matchingDataplanesForFilter(
	request *restful.Request,
	descriptor core_model.ResourceTypeDescriptor,
	resManager manager.ResourceManager,
	resourceAccess access.ResourceAccess,
	dpFilterForResource func(resource core_model.Resource) store.ListFilterFunc,
) (any, error) {
	policyName := request.PathParameter("name")
	page, err := pagination(request)
	if err != nil {
		return nil, withTitle(err, "Could not retrieve policy")
	}
	nameContains := request.QueryParameter("name")
	meshName := request.PathParameter("mesh")

	if err := resourceAccess.ValidateGet(
		request.Request.Context(),
		core_model.ResourceKey{Mesh: meshName, Name: policyName},
		descriptor,
		user.FromCtx(request.Request.Context()),
	); err != nil {
		return nil, withTitle(err, "Access Denied")
	}
	policyResource := descriptor.NewObject()
	if err := resManager.Get(request.Request.Context(), policyResource, store.GetByKey(policyName, meshName)); err != nil {
		return nil, withTitle(err, "Could not retrieve policy")
	}

	dppList := registry.Global().MustNewList(core_mesh.DataplaneType)
	err = resManager.List(request.Request.Context(), dppList,
		store.ListByMesh(meshName),
		store.ListByNameContains(nameContains),
		store.ListByFilterFunc(dpFilterForResource(policyResource)),
		store.ListByPage(page.size, page.offset),
	)
	if err != nil {
		return nil, withTitle(err, "failed inspect")
	}
	items := make([]api_common.Meta, len(dppList.GetItems()))
	for i, elt := range dppList.GetItems() {
		items[i] = oapi_helpers.ResourceToMeta(elt)
	}
	return api_types.InspectDataplanesForPolicyResponse{
		Total: int(dppList.GetPagination().Total),
		Items: items,
		Next:  nextLink(request, dppList.GetPagination().NextOffset),
	}, nil
}

func (r *resourceInspectHandler) configForProxy() handlerFunc {
	return func(request *restful.Request) (any, error) {
		ctx := request.Request.Context()

		name := request.PathParameter("name")
		mesh, err := r.meshFromRequest(request)
		if err != nil {
			return nil, withTitle(err, "Failed to retrieve Mesh")
		}
		qparams, err := r.configForProxyParams(request)
		if err != nil {
			return nil, withTitle(err, "Failed to parse query parameters")
		}

		mc, err := r.meshContextBuilder.Build(ctx, mesh)
		if err != nil {
			return nil, withTitle(err, "Failed to build mesh context")
		}

		dataplaneInsight := core_mesh.NewDataplaneInsightResource()
		err = r.resManager.Get(ctx, dataplaneInsight, store.GetByKey(name, mesh))
		if err != nil {
			return nil, withTitle(err, "Failed to fetch dataplane insight")
		}

		inspector, err := inspect.NewProxyConfigInspector(mc, core_xds.DataplaneMetadataFromXdsMetadata(dataplaneInsight.Spec.Metadata), r.zoneName, r.knownInternalAddresses, r.xdsHooks...)
		if err != nil {
			return nil, withTitle(err, "Failed to create proxy config inspector")
		}

		config, err := inspector.Get(ctx, name, *qparams.Shadow)
		if err != nil {
			return nil, withTitle(err, "Failed to inspect proxy config")
		}

		out := &api_types.GetDataplaneXDSConfigResponse{
			Xds: config,
		}

		if slices.Contains(*qparams.Include, api_types.Diff) {
			currentConfig, err := inspector.Get(ctx, name, false)
			if err != nil {
				return nil, withTitle(err, "Failed to inspect current proxy config")
			}
			diff, err := inspect.Diff(currentConfig, config)
			if err != nil {
				return nil, withTitle(err, "Failed to compute diff")
			}
			out.Diff = &diff
		}

		return out, nil
	}
}

func (r *resourceInspectHandler) configForProxyParams(request *restful.Request) (*api_types.GetDataplanesXdsConfigParams, error) {
	params := &api_types.GetDataplanesXdsConfigParams{
		Shadow:  pointer.To(false),
		Include: &[]api_types.GetDataplanesXdsConfigParamsInclude{},
	}

	if shadow := request.QueryParameter("shadow"); shadow != "" {
		if b, err := strconv.ParseBool(shadow); err != nil {
			return nil, rest_errors.NewBadRequestError("unsupported value for query parameter 'shadow'")
		} else {
			params.Shadow = &b
		}
	}

	if include := request.QueryParameter("include"); include != "" {
		for v := range strings.SplitSeq(include, ",") {
			switch api_types.GetDataplanesXdsConfigParamsInclude(v) {
			case api_types.Diff:
			default:
				return nil, rest_errors.NewBadRequestError("unsupported value for query parameter 'include'")
			}
			*params.Include = append(*params.Include, api_types.GetDataplanesXdsConfigParamsInclude(v))
		}
	}

	return params, nil
}

func (r *resourceInspectHandler) getPoliciesConf(plugins []core_plugins.RegisteredPolicyPlugin, mapToResponse matchedPoliciesToResponse) handlerFunc {
	return func(request *restful.Request) (any, error) {
		dataplaneName := request.PathParameter("name")
		meshName, err := r.meshFromRequest(request)
		if err != nil {
			return nil, withTitle(err, "Failed to retrieve Mesh")
		}

		if err := r.resourceAccess.ValidateGet(
			request.Request.Context(),
			core_model.ResourceKey{Mesh: meshName, Name: dataplaneName},
			r.descriptor,
			user.FromCtx(request.Request.Context()),
		); err != nil {
			return nil, withTitle(err, "Access Denied")
		}

		resource := r.descriptor.NewObject()
		if err := r.resManager.Get(request.Request.Context(), resource, store.GetByKey(dataplaneName, meshName)); err != nil {
			return nil, withTitle(err, fmt.Sprintf("Could not retrieve %s", r.descriptor.Name))
		}
		dataplane := resource.(*core_mesh.DataplaneResource)

		baseMeshContext, err := r.meshContextBuilder.BuildBaseMeshContextIfChanged(request.Request.Context(), meshName, nil)
		if err != nil {
			return nil, withTitle(err, "Failed to build Mesh context")
		}

		var matchedPolicies []core_xds.TypedMatchingPolicies
		allPlugins := plugins
		for _, policyPlugin := range allPlugins {
			res, err := policyPlugin.Plugin.MatchedPolicies(dataplane, baseMeshContext.Resources())
			if err != nil {
				return nil, withTitle(err, fmt.Sprintf("could not apply policy plugin %s", policyPlugin.Name))
			}
			if res.Type == "" {
				return nil, withTitle(fmt.Errorf("matched policy didn't set type for policy plugin %s", policyPlugin.Name), "could not apply policy plugin")
			}

			matchedPolicies = append(matchedPolicies, res)
		}

		out, err := mapToResponse(matchedPolicies, request, baseMeshContext.Mesh, dataplane, baseMeshContext.Resources())
		if err != nil {
			return nil, withTitle(err, "Failed building response")
		}

		return out, nil
	}
}

type matchedPoliciesToResponse func([]core_xds.TypedMatchingPolicies, *restful.Request, *core_mesh.MeshResource, *core_mesh.DataplaneResource, xds_context.Resources) (any, error)

func matchedPoliciesToProxyPolicy(matchedPolicies []core_xds.TypedMatchingPolicies, _ *restful.Request, _ *core_mesh.MeshResource, _ *core_mesh.DataplaneResource, _ xds_context.Resources) (any, error) {
	conf := []api_common.PolicyConf{}
	for _, matched := range matchedPolicies {
		if matched.ProxyConf == nil {
			continue
		}
		conf = append(conf, api_common.PolicyConf{
			Conf:    matched.ProxyConf.Conf,
			Kind:    string(matched.Type),
			Origins: policyOriginsToKRIOrigins(matched.Type, matched.ProxyConf.Origin),
		})
	}
	return api_common.PoliciesList{Policies: conf}, nil
}

func matchedPoliciesToOutboundPolicy(matchedPolicies []core_xds.TypedMatchingPolicies, request *restful.Request, mesh *core_mesh.MeshResource, _ *core_mesh.DataplaneResource, _ xds_context.Resources) (any, error) {
	outboundKri, err := kri.FromString(request.PathParameter("outbound_kri"))
	if err != nil {
		return nil, rest_errors.NewBadRequestError(err.Error())
	}

	conf := []api_common.PolicyConf{}
	for _, matched := range matchedPolicies {
		rctx := outbound.RootContext[any](mesh, matched.ToRules.ResourceRules).
			WithID(kri.NoSectionName(outboundKri)).
			WithID(outboundKri)
		computed := rctx.ResourceRule()
		if computed == nil {
			continue
		}
		conf = append(conf, api_common.PolicyConf{
			Conf:    computed.Conf,
			Kind:    string(matched.Type),
			Origins: policyOriginsToKRIOrigins(matched.Type, util_slices.Map(computed.Origin, func(o common.Origin) core_model.ResourceMeta { return o.Resource })),
		})
	}

	return api_common.PoliciesList{Policies: conf}, nil
}

func matchedPoliciesToInboundConfig(matchedPolicies []core_xds.TypedMatchingPolicies, request *restful.Request, _ *core_mesh.MeshResource, dataplane *core_mesh.DataplaneResource, resources xds_context.Resources) (any, error) {
	inboundKri, err := kri.FromString(request.PathParameter("inbound_kri"))
	if err != nil {
		return nil, rest_errors.NewBadRequestError(err.Error())
	}
	var inboundKey core_rules.InboundListener
	if inbounds := dataplane.Spec.GetNetworking().InboundsSelectedBySectionName(inboundKri.SectionName); len(inbounds) > 0 {
		inboundKey = core_rules.InboundListener{
			Address: inbounds[0].DataplaneIP,
			Port:    inbounds[0].DataplanePort,
		}
	} else if listeners := dataplane.Spec.GetNetworking().ListenersSelectedBySectionName(inboundKri.SectionName); len(listeners) > 0 {
		addr := listeners[0].GetAddress()
		if addr == "" {
			addr = dataplane.Spec.GetNetworking().GetAddress()
		}
		inboundKey = core_rules.InboundListener{Address: addr, Port: listeners[0].GetPort()}
	} else {
		return nil, errors.New("inbound not found")
	}

	conf := []api_common.InboundPolicyConf{}
	for _, matched := range matchedPolicies {
		rules := matched.FromRules.InboundRules[inboundKey]
		if len(rules) == 0 {
			continue
		}

		var policyRules []api_common.PolicyRule
		for _, rule := range rules {
			policyRules = append(policyRules, api_common.PolicyRule{
				Kri:  pointer.To(originToKRI(rule.Origin.Resource, matched.Type).Kri),
				Conf: rule.Conf,
			})
		}

		seen := map[core_model.ResourceKey]struct{}{}
		var originResources []core_model.ResourceMeta
		for _, rule := range rules {
			key := core_model.MetaToResourceKey(rule.Origin.Resource)
			if _, ok := seen[key]; !ok {
				seen[key] = struct{}{}
				originResources = append(originResources, rule.Origin.Resource)
			}
		}

		conf = append(conf, api_common.InboundPolicyConf{
			Rules:   policyRules,
			Kind:    string(matched.Type),
			Origins: policyOriginsToKRIOrigins(matched.Type, originResources),
		})
	}

	return api_common.InboundPoliciesList{Policies: conf}, nil
}

func matchedPoliciesToRoutes(matchedPolicies []core_xds.TypedMatchingPolicies, request *restful.Request, _ *core_mesh.MeshResource, _ *core_mesh.DataplaneResource, resources xds_context.Resources) (any, error) {
	outboundKri, err := kri.FromString(request.PathParameter("outbound_kri"))
	if err != nil {
		return nil, rest_errors.NewBadRequestError(err.Error())
	}

	routeConfs := []api_common.RouteConf{}
	for _, matched := range matchedPolicies {
		conf := matched.ToRules.ResourceRules.Compute(outboundKri, resources)
		if conf == nil {
			continue
		}

		var apiRules []api_common.RouteRules
		switch matched.Type {
		case meshhttproute_api.MeshHTTPRouteType:
			for _, rule := range conf.Conf {
				for _, rr := range rule.(meshhttproute_api.PolicyDefault).Rules {
					apiRules = append(apiRules, api_common.RouteRules{
						Conf:    rr.Default,
						Kri:     originToKRI(conf.OriginByMatches[meshhttproute_api.HashMatches(rr.Matches)].Resource, matched.Type).Kri,
						Matches: util_slices.Map(rr.Matches, func(m meshhttproute_api.Match) any { return m }),
					})
				}
			}
		case meshtcproute_api.MeshTCPRouteType:
			for _, rule := range conf.Conf {
				apiRules = append(apiRules, api_common.RouteRules{Conf: rule.(meshtcproute_api.Rule).Default})
			}
		}

		routeConfs = append(routeConfs, api_common.RouteConf{
			Kind:    string(matched.Type),
			Rules:   apiRules,
			Origins: policyOriginsToKRIOrigins(matched.Type, util_slices.Map(conf.Origin, func(o common.Origin) core_model.ResourceMeta { return o.Resource })),
		})
	}

	return api_common.RoutesList{Routes: routeConfs}, nil
}

func matchedPoliciesToRouteConfig(matchedPolicies []core_xds.TypedMatchingPolicies, request *restful.Request, mesh *core_mesh.MeshResource, _ *core_mesh.DataplaneResource, resources xds_context.Resources) (any, error) {
	outboundKri, err := kri.FromString(request.PathParameter("outbound_kri"))
	if err != nil {
		return nil, rest_errors.NewBadRequestError(err.Error())
	}
	routeKri, err := kri.FromString(request.PathParameter("route_kri"))
	if err != nil {
		return nil, rest_errors.NewBadRequestError(err.Error())
	}

	conf := []api_common.PolicyConf{}
	for _, matched := range matchedPolicies {
		rctx := outbound.RootContext[any](mesh, matched.ToRules.ResourceRules).
			WithID(kri.NoSectionName(outboundKri)).
			WithID(outboundKri).
			WithID(routeKri)
		computed := rctx.ResourceRule()
		if computed == nil {
			continue
		}

		conf = append(conf, api_common.PolicyConf{
			Conf:    computed.Conf,
			Kind:    string(matched.Type),
			Origins: policyOriginsToKRIOrigins(matched.Type, util_slices.Map(computed.Origin, func(o common.Origin) core_model.ResourceMeta { return o.Resource })),
		})
	}

	return api_common.PoliciesList{Policies: conf}, nil
}

func policyOriginsToKRIOrigins(policyType core_model.ResourceType, origins []core_model.ResourceMeta) []api_common.PolicyOrigin {
	return util_slices.Map(origins, func(origin core_model.ResourceMeta) api_common.PolicyOrigin {
		return originToKRI(origin, policyType)
	})
}

func originToKRI(origin core_model.ResourceMeta, policyType core_model.ResourceType) api_common.PolicyOrigin {
	return api_common.PolicyOrigin{Kri: kri.FromResourceMeta(origin, policyType).String()}
}

func (r *resourceInspectHandler) rulesForResource() handlerFunc {
	return func(request *restful.Request) (any, error) {
		resourceName := request.PathParameter("name")
		meshName, err := r.meshFromRequest(request)
		if err != nil {
			return nil, withTitle(err, "Failed to retrieve Mesh")
		}

		if err := r.resourceAccess.ValidateGet(
			request.Request.Context(),
			core_model.ResourceKey{Mesh: meshName, Name: resourceName},
			r.descriptor,
			user.FromCtx(request.Request.Context()),
		); err != nil {
			return nil, withTitle(err, "Access Denied")
		}

		resource := r.descriptor.NewObject()
		if err := r.resManager.Get(request.Request.Context(), resource, store.GetByKey(resourceName, meshName)); err != nil {
			return nil, withTitle(err, fmt.Sprintf("Could not retrieve %s", r.descriptor.Name))
		}
		var dp *core_mesh.DataplaneResource
		switch r.descriptor.Name {
		case core_mesh.DataplaneType:
			dp = resource.(*core_mesh.DataplaneResource)
		default:
			return nil, withTitle(fmt.Errorf("rules not supported for type %s", r.descriptor.Name), "Unsupported resource type")
		}
		baseMeshContext, err := r.meshContextBuilder.BuildBaseMeshContextIfChanged(request.Request.Context(), meshName, nil)
		if err != nil {
			return nil, withTitle(err, "Failed to build Mesh context")
		}

		resources := xds_context.Resources{
			MeshLocalResources: baseMeshContext.ResourceMap,
		}
		matchesByHash := map[common_api.MatchesHash][]meshhttproute_api.Match{}
		// Get all the matching policies
		allPlugins := core_plugins.Plugins().PolicyPlugins()
		rules := []api_common.InspectRule{}
		for _, policyPlugin := range allPlugins {
			res, err := policyPlugin.Plugin.MatchedPolicies(dp, resources)
			if err != nil {
				return nil, withTitle(err, fmt.Sprintf("could not apply policy plugin %s", policyPlugin.Name))
			}
			if res.Type == "" {
				return nil, withTitle(fmt.Errorf("matched policy didn't set type for policy plugin %s", policyPlugin.Name), "could not apply policy plugin")
			}
			if res.Type == meshhttproute_api.MeshHTTPRouteType {
				for _, resourceRule := range res.ToRules.ResourceRules {
					for _, conf := range resourceRule.Conf {
						if pd, ok := conf.(meshhttproute_api.PolicyDefault); ok {
							for _, r := range pd.Rules {
								matchesByHash[meshhttproute_api.HashMatches(r.Matches)] = r.Matches
							}
						}
					}
				}
			}

			if len(res.ToRules.ResourceRules) == 0 && len(res.FromRules.InboundRules) == 0 && res.ProxyConf == nil {
				continue
			}
			var proxyRule *api_common.ProxyRule
			if res.ProxyConf != nil {
				proxyRule = &api_common.ProxyRule{
					Conf:   res.ProxyConf.Conf,
					Origin: oapi_helpers.ResourceMetaListToMetaList(res.Type, res.ProxyConf.Origin),
				}
			}

			getInboundPortName := func(port uint32) *string {
				if name := dp.Spec.GetNetworking().GetInboundForPort(port).GetName(); name != "" {
					return &name
				}
				return nil
			}

			inboundRules := []api_common.InboundRulesEntry{}
			for inbound, rulesForInbound := range res.FromRules.InboundRules {
				if len(rulesForInbound) == 0 {
					continue
				}
				rs := make([]api_common.InboundRule, len(rulesForInbound))
				for i := range rulesForInbound {
					rs[i] = api_common.InboundRule{
						Conf:   []any{rulesForInbound[i].Conf},
						Match:  rulesForInbound[i].Match,
						Origin: oapi_helpers.OriginListToResourceRuleOrigin(res.Type, []common.Origin{rulesForInbound[i].Origin}),
					}
				}
				inboundRules = append(inboundRules, api_common.InboundRulesEntry{
					Inbound: api_common.Inbound{
						Name: getInboundPortName(inbound.Port),
						Port: int(inbound.Port),
					},
					Rules: rs,
				})
			}
			sort.SliceStable(inboundRules, func(i, j int) bool {
				return inboundRules[i].Inbound.Port < inboundRules[j].Inbound.Port
			})

			toResourceRules := []api_common.ResourceRule{}
			for itemIdentifier, resourceRuleItem := range res.ToRules.ResourceRules {
				toResourceRules = append(toResourceRules, api_common.ResourceRule{
					Conf:                resourceRuleItem.Conf,
					Origin:              oapi_helpers.OriginListToResourceRuleOrigin(res.Type, resourceRuleItem.Origin),
					ResourceMeta:        oapi_helpers.ResourceMetaToMeta(itemIdentifier.ResourceType, resourceRuleItem.Resource),
					ResourceSectionName: &resourceRuleItem.ResourceSectionName,
				})
			}
			sort.Slice(toResourceRules, func(i, j int) bool {
				return toResourceRules[i].ResourceMeta.Name < toResourceRules[j].ResourceMeta.Name
			})

			if proxyRule == nil && len(toResourceRules) == 0 && len(inboundRules) == 0 && len(res.Warnings) == 0 {
				// No matches for this policy, keep going...
				continue
			}
			warnings := res.Warnings
			if warnings == nil {
				warnings = []string{}
			}
			rules = append(rules, api_common.InspectRule{
				Type:            string(res.Type),
				ToResourceRules: &toResourceRules,
				InboundRules:    &inboundRules,
				ProxyRule:       proxyRule,
				Warnings:        &warnings,
			})
		}
		httpMatches := []api_common.HttpMatch{}
		for k, v := range matchesByHash {
			httpMatches = append(httpMatches, api_common.HttpMatch{
				Match: v,
				Hash:  string(k),
			})
		}
		sort.Slice(httpMatches, func(i, j int) bool {
			return httpMatches[i].Hash < httpMatches[j].Hash
		})
		out := api_types.InspectRulesResponse{
			HttpMatches: httpMatches,
			Resource:    oapi_helpers.ResourceToMeta(resource),
			Rules:       rules,
		}
		return out, nil
	}
}
