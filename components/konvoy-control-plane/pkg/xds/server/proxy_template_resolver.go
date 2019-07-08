package server

import (
	"context"

	mesh_proto "github.com/Kong/konvoy/components/konvoy-control-plane/api/mesh/v1alpha1"
	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core"
	mesh_core "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/resources/apis/mesh"
	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/resources/store"
	mesh_k8s "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/plugins/resources/k8s/native/api/v1alpha1"
	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/xds/model"
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
	if proxy.Workload.Meta.Labels != nil {
		if templateName := proxy.Workload.Meta.Labels[mesh_k8s.ProxyTemplateAnnotation]; templateName != "" {
			template := &mesh_core.ProxyTemplateResource{}
			if err := r.ResourceStore.Get(context.Background(), template,
				store.GetByName(proxy.Workload.Meta.Namespace, templateName)); err != nil {
				templateResolverLog.Error(err, "failed to resolve ProxyTemplate",
					"workloadNamespace", proxy.Workload.Meta.Namespace,
					"workloadName", proxy.Workload.Meta.Name,
					"templateName", templateName)
			} else {
				templateResolverLog.V(1).Info("resolved ProxyTemplate",
					"workloadNamespace", proxy.Workload.Meta.Namespace,
					"workloadName", proxy.Workload.Meta.Name,
					"templateName", templateName,
				)
				return &template.Spec
			}
		}
	}
	return r.DefaultProxyTemplate
}
