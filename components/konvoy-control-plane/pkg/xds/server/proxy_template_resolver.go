package server

import (
	"context"

	mesh_proto "github.com/Kong/konvoy/components/konvoy-control-plane/api/mesh/v1alpha1"
	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core"
	mesh_core "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/resources/apis/mesh"
	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/resources/store"
	model "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/xds"
)

var (
	templateResolverLog = core.Log.WithName("proxy-template-resolver")
)

type proxyTemplateResolver interface {
	GetTemplate(proxy *model.Proxy) *mesh_proto.ProxyTemplate
}

type simpleProxyTemplateResolver struct {
	store.ResourceStore
	DefaultProxyTemplate *mesh_proto.ProxyTemplate
}

func (r *simpleProxyTemplateResolver) GetTemplate(proxy *model.Proxy) *mesh_proto.ProxyTemplate {
	ctx := context.Background()
	templateList := &mesh_core.ProxyTemplateResourceList{}
	if err := r.ResourceStore.List(ctx, templateList); err != nil {
		templateResolverLog.Error(err, "failed to resolve ProxyTemplate")
	} else if 0 < len(templateList.Items) {
		// TODO(yskopets): Use ProxyTemplate's selector to pick a proper one
		return &templateList.Items[0].Spec
	}
	return r.DefaultProxyTemplate
}
