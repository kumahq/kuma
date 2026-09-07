package api_server

import (
	"context"
	"errors"
	"net/http"
	"slices"
	"testing"

	. "github.com/onsi/gomega"

	core_plugins "github.com/kumahq/kuma/v3/pkg/core/plugins"
	"github.com/kumahq/kuma/v3/pkg/core/resources/access"
	core_mesh "github.com/kumahq/kuma/v3/pkg/core/resources/apis/mesh"
	meshservice_api "github.com/kumahq/kuma/v3/pkg/core/resources/apis/meshservice/api/v1alpha1"
	"github.com/kumahq/kuma/v3/pkg/core/resources/manager"
	core_model "github.com/kumahq/kuma/v3/pkg/core/resources/model"
	"github.com/kumahq/kuma/v3/pkg/core/resources/store"
	core_user "github.com/kumahq/kuma/v3/pkg/core/user"
	core_xds "github.com/kumahq/kuma/v3/pkg/core/xds"
	core_rules "github.com/kumahq/kuma/v3/pkg/plugins/policies/core/rules"
	rules_common "github.com/kumahq/kuma/v3/pkg/plugins/policies/core/rules/common"
	"github.com/kumahq/kuma/v3/pkg/plugins/policies/core/rules/inbound"
	"github.com/kumahq/kuma/v3/pkg/plugins/policies/core/rules/outbound"
	meshhttproute_api "github.com/kumahq/kuma/v3/pkg/plugins/policies/meshhttproute/api/v1alpha1"
	"github.com/kumahq/kuma/v3/pkg/test/resources/builders"
	test_model "github.com/kumahq/kuma/v3/pkg/test/resources/model"
	xds_context "github.com/kumahq/kuma/v3/pkg/xds/context"
)

func TestLoadDataplaneForInspection(t *testing.T) {
	t.Run("loads Mesh, authorizes, loads Dataplane, and builds context in order", func(t *testing.T) {
		g := NewWithT(t)
		events := []string{}
		dataplane := builders.Dataplane().WithName("dp-1").WithMesh("default").Build()
		baseMeshContext := &xds_context.BaseMeshContext{
			Mesh:        builders.Mesh().WithName("default").Build(),
			ResourceMap: xds_context.ResourceMap{},
		}
		handler := newInspectionHandler(
			&inspectionResourceManager{events: &events, dataplane: dataplane},
			&inspectionResourceAccess{events: &events},
			&inspectionMeshContextBuilder{events: &events, baseMeshContext: baseMeshContext},
		)
		request := newCrudRequest(http.MethodGet, "/meshes/default/dataplanes/dp-1/_rules", "dp-1", "")
		request.PathParameters()["mesh"] = "default"

		loadedDataplane, loadedContext, err := handler.loadDataplaneForInspection(request)

		g.Expect(err).NotTo(HaveOccurred())
		g.Expect(loadedDataplane.GetMeta()).To(Equal(dataplane.GetMeta()))
		g.Expect(loadedContext).To(BeIdenticalTo(baseMeshContext))
		g.Expect(events).To(Equal([]string{"mesh", "authorize", "dataplane", "context"}))
	})

	t.Run("authorization failure stops before Dataplane and context loading", func(t *testing.T) {
		g := NewWithT(t)
		events := []string{}
		accessErr := errors.New("access denied")
		handler := newInspectionHandler(
			&inspectionResourceManager{events: &events},
			&inspectionResourceAccess{events: &events, err: accessErr},
			&inspectionMeshContextBuilder{events: &events},
		)
		request := newCrudRequest(http.MethodGet, "/meshes/default/dataplanes/dp-1/_rules", "dp-1", "")
		request.PathParameters()["mesh"] = "default"

		_, _, err := handler.loadDataplaneForInspection(request)

		expectTitledError(g, err, "Access Denied")
		g.Expect(errors.Is(err, accessErr)).To(BeTrue())
		g.Expect(events).To(Equal([]string{"mesh", "authorize"}))
	})
}

func TestMatchPolicies(t *testing.T) {
	t.Run("preserves plugin order and inputs", func(t *testing.T) {
		g := NewWithT(t)
		events := []string{}
		dataplane := builders.Dataplane().Build()
		resources := xds_context.NewResources()
		plugins := []core_plugins.RegisteredPolicyPlugin{
			inspectionRegisteredPlugin("z-plugin", "MeshZ", &events),
			inspectionRegisteredPlugin("a-plugin", "MeshA", &events),
		}

		matched, err := matchPolicies(plugins, dataplane, resources)

		g.Expect(err).NotTo(HaveOccurred())
		g.Expect(events).To(Equal([]string{"z-plugin", "a-plugin"}))
		g.Expect(matched).To(HaveLen(2))
		g.Expect(matched[0].Type).To(Equal(core_model.ResourceType("MeshZ")))
		g.Expect(matched[1].Type).To(Equal(core_model.ResourceType("MeshA")))
		for _, plugin := range plugins {
			recording := plugin.Plugin.(*inspectionPolicyPlugin)
			g.Expect(recording.dataplane).To(BeIdenticalTo(dataplane))
			g.Expect(recording.resources).To(Equal(resources))
			g.Expect(recording.options).To(BeZero())
		}
	})

	t.Run("stops on the first plugin error", func(t *testing.T) {
		g := NewWithT(t)
		events := []string{}
		pluginErr := errors.New("match failed")
		plugins := []core_plugins.RegisteredPolicyPlugin{
			{
				Name: "broken",
				Plugin: &inspectionPolicyPlugin{
					name:   "broken",
					events: &events,
					err:    pluginErr,
				},
			},
			inspectionRegisteredPlugin("not-called", "MeshA", &events),
		}

		_, err := matchPolicies(plugins, builders.Dataplane().Build(), xds_context.NewResources())

		expectTitledError(g, err, "could not apply policy plugin broken")
		g.Expect(errors.Is(err, pluginErr)).To(BeTrue())
		g.Expect(events).To(Equal([]string{"broken"}))
	})

	t.Run("rejects an empty policy type before calling the next plugin", func(t *testing.T) {
		g := NewWithT(t)
		events := []string{}
		plugins := []core_plugins.RegisteredPolicyPlugin{
			inspectionRegisteredPlugin("empty", "", &events),
			inspectionRegisteredPlugin("not-called", "MeshA", &events),
		}

		_, err := matchPolicies(plugins, builders.Dataplane().Build(), xds_context.NewResources())

		expectTitledError(g, err, "could not apply policy plugin")
		g.Expect(err.Error()).To(ContainSubstring("matched policy didn't set type for policy plugin empty"))
		g.Expect(events).To(Equal([]string{"empty"}))
	})
}

func TestMatchedPoliciesToRulesResponse(t *testing.T) {
	g := NewWithT(t)
	dataplane := builders.Dataplane().
		WithName("dp-1").
		WithoutInbounds().
		AddInbound(builders.Inbound().WithPort(9090).WithName("inbound-9090")).
		AddInbound(builders.Inbound().WithPort(7070).WithName("inbound-7070")).
		AddInbound(builders.Inbound().WithPort(8080).WithName("inbound-8080")).
		Build()
	policyMeta := &test_model.ResourceMeta{Name: "route-policy", Mesh: "default"}
	matchesA := []meshhttproute_api.Match{{Path: &meshhttproute_api.PathMatch{Type: meshhttproute_api.Exact, Value: "/a"}}}
	matchesZ := []meshhttproute_api.Match{{Path: &meshhttproute_api.PathMatch{Type: meshhttproute_api.Exact, Value: "/z"}}}
	matched := []core_xds.TypedMatchingPolicies{{
		Type: meshhttproute_api.MeshHTTPRouteType,
		FromRules: core_rules.FromRules{InboundRules: map[core_rules.InboundListener][]*inbound.Rule{
			{Port: 9090}: {{Conf: "9090", Origin: rules_common.Origin{Resource: policyMeta}}},
			{Port: 7070}: {{Conf: "7070", Origin: rules_common.Origin{Resource: policyMeta}}},
			{Port: 8080}: {{Conf: "8080", Origin: rules_common.Origin{Resource: policyMeta}}},
		}},
		ToRules: core_rules.ToRules{ResourceRules: outbound.ResourceRules{
			{ResourceType: meshservice_api.MeshServiceType, Mesh: "default", Name: "backend-2"}: {
				Resource: &test_model.ResourceMeta{Name: "backend-2", Mesh: "default"},
				Conf: []any{meshhttproute_api.PolicyDefault{Rules: []meshhttproute_api.Rule{{
					Matches: matchesZ,
				}}}},
			},
			{ResourceType: meshservice_api.MeshServiceType, Mesh: "default", Name: "backend-1"}: {
				Resource: &test_model.ResourceMeta{Name: "backend-1", Mesh: "default"},
				Conf: []any{meshhttproute_api.PolicyDefault{Rules: []meshhttproute_api.Rule{{
					Matches: matchesA,
				}}}},
			},
		}},
	}}

	response := matchedPoliciesToRulesResponse(matched, dataplane)

	g.Expect(response.Resource.Name).To(Equal("dp-1"))
	g.Expect(response.Rules).To(HaveLen(1))
	rule := response.Rules[0]
	g.Expect(rule.InboundRules).NotTo(BeNil())
	g.Expect(*rule.InboundRules).To(HaveLen(3))
	g.Expect([]int{
		(*rule.InboundRules)[0].Inbound.Port,
		(*rule.InboundRules)[1].Inbound.Port,
		(*rule.InboundRules)[2].Inbound.Port,
	}).To(Equal([]int{7070, 8080, 9090}))
	g.Expect(rule.ToResourceRules).NotTo(BeNil())
	g.Expect(*rule.ToResourceRules).To(HaveLen(2))
	g.Expect([]string{
		(*rule.ToResourceRules)[0].ResourceMeta.Name,
		(*rule.ToResourceRules)[1].ResourceMeta.Name,
	}).To(Equal([]string{"backend-1", "backend-2"}))
	g.Expect(rule.Warnings).NotTo(BeNil())
	g.Expect(*rule.Warnings).To(BeEmpty())
	g.Expect(response.HttpMatches).To(HaveLen(2))
	hashes := []string{response.HttpMatches[0].Hash, response.HttpMatches[1].Hash}
	expectedHashes := []string{string(meshhttproute_api.HashMatches(matchesA)), string(meshhttproute_api.HashMatches(matchesZ))}
	slices.Sort(expectedHashes)
	g.Expect(hashes).To(Equal(expectedHashes))

	emptyResponse := matchedPoliciesToRulesResponse(nil, dataplane)
	g.Expect(emptyResponse.Rules).NotTo(BeNil())
	g.Expect(emptyResponse.Rules).To(BeEmpty())
	g.Expect(emptyResponse.HttpMatches).NotTo(BeNil())
	g.Expect(emptyResponse.HttpMatches).To(BeEmpty())
}

type inspectionResourceManager struct {
	manager.ResourceManager
	events    *[]string
	dataplane *core_mesh.DataplaneResource
}

func (r *inspectionResourceManager) Get(_ context.Context, resource core_model.Resource, _ ...store.GetOptionsFunc) error {
	switch typed := resource.(type) {
	case *core_mesh.MeshResource:
		*r.events = append(*r.events, "mesh")
	case *core_mesh.DataplaneResource:
		*r.events = append(*r.events, "dataplane")
		if r.dataplane != nil {
			typed.SetMeta(r.dataplane.GetMeta())
			return typed.SetSpec(r.dataplane.GetSpec())
		}
	}
	return nil
}

type inspectionResourceAccess struct {
	access.ResourceAccess
	events *[]string
	err    error
}

func (r *inspectionResourceAccess) ValidateGet(context.Context, core_model.ResourceKey, core_model.ResourceTypeDescriptor, core_user.User) error {
	*r.events = append(*r.events, "authorize")
	return r.err
}

type inspectionMeshContextBuilder struct {
	xds_context.MeshContextBuilder
	events          *[]string
	baseMeshContext *xds_context.BaseMeshContext
}

func (b *inspectionMeshContextBuilder) BuildBaseMeshContextIfChanged(context.Context, string, *xds_context.BaseMeshContext) (*xds_context.BaseMeshContext, error) {
	*b.events = append(*b.events, "context")
	return b.baseMeshContext, nil
}

func newInspectionHandler(resManager manager.ResourceManager, resourceAccess access.ResourceAccess, meshContextBuilder xds_context.MeshContextBuilder) *resourceInspectHandler {
	return &resourceInspectHandler{
		resManager:         resManager,
		descriptor:         core_mesh.DataplaneResourceTypeDescriptor,
		resourceAccess:     resourceAccess,
		meshContextBuilder: meshContextBuilder,
	}
}

type inspectionPolicyPlugin struct {
	name      string
	events    *[]string
	result    core_xds.TypedMatchingPolicies
	err       error
	dataplane *core_mesh.DataplaneResource
	resources xds_context.Resources
	options   int
}

func (p *inspectionPolicyPlugin) Order() int {
	return 0
}

func (p *inspectionPolicyPlugin) MatchedPolicies(dataplane *core_mesh.DataplaneResource, resources xds_context.Resources, opts ...core_plugins.MatchedPoliciesOption) (core_xds.TypedMatchingPolicies, error) {
	*p.events = append(*p.events, p.name)
	p.dataplane = dataplane
	p.resources = resources
	p.options = len(opts)
	return p.result, p.err
}

func (p *inspectionPolicyPlugin) Apply(*core_xds.ResourceSet, xds_context.Context, *core_xds.Proxy) error {
	return nil
}

func inspectionRegisteredPlugin(name string, resourceType core_model.ResourceType, events *[]string) core_plugins.RegisteredPolicyPlugin {
	return core_plugins.RegisteredPolicyPlugin{
		Name: core_plugins.PluginName(name),
		Plugin: &inspectionPolicyPlugin{
			name:   name,
			events: events,
			result: core_xds.TypedMatchingPolicies{Type: resourceType},
		},
	}
}
