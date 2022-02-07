package template

import (
	"context"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core"
	core_policy "github.com/kumahq/kuma/pkg/core/policy"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/manager"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	core_store "github.com/kumahq/kuma/pkg/core/resources/store"
	model "github.com/kumahq/kuma/pkg/core/xds"
)

var (
	templateResolverLog = core.Log.WithName("proxy-template-resolver")
)

type ProxyTemplateResolver interface {
	GetTemplate(proxy *model.Proxy) *mesh_proto.ProxyTemplate
}

type SimpleProxyTemplateResolver struct {
	ReadOnlyResourceManager manager.ReadOnlyResourceManager
}

func (r *SimpleProxyTemplateResolver) GetTemplate(proxy *model.Proxy) *mesh_proto.ProxyTemplate {
	log := templateResolverLog.WithValues("dataplane", core_model.MetaToResourceKey(proxy.Dataplane.Meta))
	ctx := context.Background()
	templateList := &core_mesh.ProxyTemplateResourceList{}
	if err := r.ReadOnlyResourceManager.List(ctx, templateList, core_store.ListByMesh(proxy.Dataplane.Meta.GetMesh())); err != nil {
		templateResolverLog.Error(err, "failed to list ProxyTemplates")
		return nil
	}

	if bestMatchTemplate := SelectProxyTemplate(proxy.Dataplane, templateList.Items); bestMatchTemplate != nil {
		log.V(2).Info("found the best matching ProxyTemplate", "proxytemplate", core_model.MetaToResourceKey(bestMatchTemplate.GetMeta()))
		return bestMatchTemplate.Spec
	}

	log.V(2).Info("no matching ProxyTemplate")
	return nil
}

type StaticProxyTemplateResolver struct {
	Template *mesh_proto.ProxyTemplate
}

func (r *StaticProxyTemplateResolver) GetTemplate(proxy *model.Proxy) *mesh_proto.ProxyTemplate {
	return r.Template
}

type sequentialResolver []ProxyTemplateResolver

func (s sequentialResolver) GetTemplate(proxy *model.Proxy) *mesh_proto.ProxyTemplate {
	for _, r := range s {
		if t := r.GetTemplate(proxy); t != nil {
			return t
		}
	}

	return nil
}

// SequentialResolver returns a new ProxyTemplate resolver that applies
// each of the resolvers given as arguments in turn. The result of the
// first successful resolver is returned.
func SequentialResolver(r ...ProxyTemplateResolver) ProxyTemplateResolver {
	return sequentialResolver(r)
}

func SelectProxyTemplate(dataplane *core_mesh.DataplaneResource, proxyTemplates []*core_mesh.ProxyTemplateResource) *core_mesh.ProxyTemplateResource {
	policies := make([]core_policy.DataplanePolicy, len(proxyTemplates))
	for i, proxyTemplate := range proxyTemplates {
		policies[i] = proxyTemplate
	}
	if policy := core_policy.SelectDataplanePolicy(dataplane, policies); policy != nil {
		return policy.(*core_mesh.ProxyTemplateResource)
	}
	return nil
}
