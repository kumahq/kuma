package server

import (
	"context"
	"sort"

	mesh_proto "github.com/Kong/konvoy/components/konvoy-control-plane/api/mesh/v1alpha1"
	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core"
	mesh_core "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/resources/apis/mesh"
	core_model "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/resources/model"
	core_store "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/resources/store"
	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/xds/model"
)

var (
	templateResolverLog = core.Log.WithName("proxy-template-resolver")
)

type proxyTemplateResolver interface {
	GetTemplate(proxy *model.Proxy) *mesh_proto.ProxyTemplate
}

type simpleProxyTemplateResolver struct {
	core_store.ResourceStore
	DefaultProxyTemplate *mesh_proto.ProxyTemplate
}

func (r *simpleProxyTemplateResolver) GetTemplate(proxy *model.Proxy) *mesh_proto.ProxyTemplate {
	log := templateResolverLog.WithValues("dataplane", core_model.MetaToResourceKey(proxy.Dataplane.Meta))
	ctx := context.Background()
	templateList := &mesh_core.ProxyTemplateResourceList{}
	if err := r.ResourceStore.List(ctx, templateList, core_store.ListByMesh(proxy.Dataplane.Meta.GetMesh())); err != nil {
		templateResolverLog.Error(err, "failed to list ProxyTemplates")
	}
	if bestMatchTemplate := FindBestMatch(proxy, templateList.Items); bestMatchTemplate != nil {
		log.V(1).Info("found the best matching ProxyTemplate", "proxytemplate", core_model.MetaToResourceKey(bestMatchTemplate.Meta))
		return &bestMatchTemplate.Spec
	}
	log.V(1).Info("falling back to the default ProxyTemplate since there is no best match", "templates", templateList.Items)
	return r.DefaultProxyTemplate
}

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
