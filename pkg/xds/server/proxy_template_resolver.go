package server

import (
	"context"

	mesh_proto "github.com/Kong/kuma/api/mesh/v1alpha1"
	"github.com/Kong/kuma/pkg/core"
	core_policy "github.com/Kong/kuma/pkg/core/policy"
	mesh_core "github.com/Kong/kuma/pkg/core/resources/apis/mesh"
	"github.com/Kong/kuma/pkg/core/resources/manager"
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
	manager.ReadOnlyResourceManager
	DefaultProxyTemplate *mesh_proto.ProxyTemplate
}

func (r *simpleProxyTemplateResolver) GetTemplate(proxy *model.Proxy) *mesh_proto.ProxyTemplate {
	log := templateResolverLog.WithValues("dataplane", core_model.MetaToResourceKey(proxy.Dataplane.Meta))
	ctx := context.Background()
	templateList := &mesh_core.ProxyTemplateResourceList{}
	if err := r.ReadOnlyResourceManager.List(ctx, templateList, core_store.ListByMesh(proxy.Dataplane.Meta.GetMesh())); err != nil {
		templateResolverLog.Error(err, "failed to list ProxyTemplates")
		return nil
	}

	policies := make([]core_policy.DataplanePolicy, len(templateList.Items))
	for i, proxyTemplate := range templateList.Items {
		policies[i] = proxyTemplate
	}

	if bestMatchTemplate := core_policy.SelectDataplanePolicy(proxy.Dataplane, policies); bestMatchTemplate != nil {
		log.V(2).Info("found the best matching ProxyTemplate", "proxytemplate", core_model.MetaToResourceKey(bestMatchTemplate.GetMeta()))
		return &bestMatchTemplate.(*mesh_core.ProxyTemplateResource).Spec
	}
	log.V(2).Info("falling back to the default ProxyTemplate since there is no best match", "templates", templateList.Items)
	return r.DefaultProxyTemplate
}
