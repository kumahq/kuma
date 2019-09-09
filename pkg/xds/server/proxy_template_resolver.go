package server

import (
	"context"
	"github.com/Kong/kuma/pkg/core/resources/manager"
	"sort"

	mesh_proto "github.com/Kong/kuma/api/mesh/v1alpha1"
	"github.com/Kong/kuma/pkg/core"
	mesh_core "github.com/Kong/kuma/pkg/core/resources/apis/mesh"
	core_model "github.com/Kong/kuma/pkg/core/resources/model"
	core_store "github.com/Kong/kuma/pkg/core/resources/store"
	model "github.com/Kong/kuma/pkg/core/xds"
)

var (
	templateResolverLog = core.Log.WithName("proxy-template-resolver")
)

type proxyTemplateResolver interface {
	GetTemplate(proxy *model.Proxy) *mesh_proto.ProxyTemplate
}

type simpleProxyTemplateResolver struct {
	manager.ResourceManager
	DefaultProxyTemplate *mesh_proto.ProxyTemplate
}

func (r *simpleProxyTemplateResolver) GetTemplate(proxy *model.Proxy) *mesh_proto.ProxyTemplate {
	log := templateResolverLog.WithValues("dataplane", core_model.MetaToResourceKey(proxy.Dataplane.Meta))
	ctx := context.Background()
	templateList := &mesh_core.ProxyTemplateResourceList{}
	if err := r.ResourceManager.List(ctx, templateList, core_store.ListByMesh(proxy.Dataplane.Meta.GetMesh())); err != nil {
		templateResolverLog.Error(err, "failed to list ProxyTemplates")
	}
	if bestMatchTemplate := FindBestMatch(proxy, templateList.Items); bestMatchTemplate != nil {
		log.V(2).Info("found the best matching ProxyTemplate", "proxytemplate", core_model.MetaToResourceKey(bestMatchTemplate.Meta))
		return &bestMatchTemplate.Spec
	}
	log.V(2).Info("falling back to the default ProxyTemplate since there is no best match", "templates", templateList.Items)
	return r.DefaultProxyTemplate
}

// FindBestMatch given a Dataplane definition and a list of ProxyTemplates returns the "best matching" ProxyTemplate.
// A ProxyTemplate is considered a match if one of the inbound interfaces of a Dataplane has all tags of ProxyTemplate's selector.
// Every matching ProxyTemplate gets a rank (score) defined as a maximum number of tags in a matching selector.
// ProxyTemplate with an empty list of selectors is considered a match with a rank (score) of 0.
// ProxyTemplate with an empty selector (one that has no tags) is considered a match with a rank (score) of 0.
// In case if there are multiple ProxyTemplates with the same rank (score), templates are sorted alphabetically by Namespace and Name
// and the first one is considered the "best match".
func FindBestMatch(proxy *model.Proxy, templates []*mesh_core.ProxyTemplateResource) *mesh_core.ProxyTemplateResource {
	sort.Stable(ProxyTemplatesByNamespacedName(templates)) // sort to avoid flakiness

	var bestMatch *mesh_core.ProxyTemplateResource
	var bestScore int
	for _, template := range templates {
		if 0 == len(template.Spec.Selectors) { // match everything
			if bestMatch == nil {
				bestMatch = template
			}
			continue
		}
		for _, selector := range template.Spec.Selectors {
			if 0 == len(selector.Match) { // match everything
				if bestMatch == nil {
					bestMatch = template
				}
				continue
			}
			for _, inbound := range proxy.Dataplane.Spec.Networking.GetInbound() {
				if matches, score := ScoreMatch(selector.Match, inbound.Tags); matches && bestScore < score {
					bestMatch = template
					bestScore = score
				}
			}
		}
	}
	return bestMatch
}

func ScoreMatch(selector map[string]string, target map[string]string) (bool, int) {
	for key, requiredValue := range selector {
		if actualValue, hasKey := target[key]; !hasKey || actualValue != requiredValue {
			return false, 0
		}
	}
	return true, len(selector)
}

type ProxyTemplatesByNamespacedName []*mesh_core.ProxyTemplateResource

func (a ProxyTemplatesByNamespacedName) Len() int      { return len(a) }
func (a ProxyTemplatesByNamespacedName) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a ProxyTemplatesByNamespacedName) Less(i, j int) bool {
	return a[i].Meta.GetNamespace() < a[j].Meta.GetNamespace() ||
		(a[i].Meta.GetNamespace() == a[j].Meta.GetNamespace() && a[i].Meta.GetName() < a[j].Meta.GetName())
}
