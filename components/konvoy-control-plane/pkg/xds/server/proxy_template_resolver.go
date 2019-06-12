package server

import (
	"context"

	konvoy_mesh "github.com/Kong/konvoy/components/konvoy-control-plane/model/api/v1alpha1"
	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/xds/model"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	templateResolverLog = ctrl.Log.WithName("proxy-template-resolver")
)

type proxyTemplateResolver interface {
	GetTemplate(proxy *model.Proxy) *konvoy_mesh.ProxyTemplate
}

type simpleProxyTemplateResolver struct {
	client.Client
	DefaultProxyTemplate *konvoy_mesh.ProxyTemplate
}

func (r *simpleProxyTemplateResolver) GetTemplate(proxy *model.Proxy) *konvoy_mesh.ProxyTemplate {
	if proxy.Workload.Meta != nil && proxy.Workload.Meta.GetAnnotations() != nil {
		if templateName := proxy.Workload.Meta.GetAnnotations()[konvoy_mesh.ProxyTemplateAnnotation]; templateName != "" {
			template := &konvoy_mesh.ProxyTemplate{}
			if err := r.Client.Get(
				context.Background(),
				types.NamespacedName{Namespace: proxy.Workload.Meta.GetNamespace(), Name: templateName},
				template); err != nil {
				templateResolverLog.Error(err, "failed to resolve ProxyTemplate",
					"workloadNamespace", proxy.Workload.Meta.GetNamespace(),
					"workloadName", proxy.Workload.Meta.GetName(),
					"templateName", templateName,
				)
			} else {
				templateResolverLog.V(1).Info("resolved ProxyTemplate",
					"workloadNamespace", proxy.Workload.Meta.GetNamespace(),
					"workloadName", proxy.Workload.Meta.GetName(),
					"templateName", templateName,
				)
				return template
			}
		}
	}
	return r.DefaultProxyTemplate
}
