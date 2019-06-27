package server

import (
	"context"

	konvoy_mesh "github.com/Kong/konvoy/components/konvoy-control-plane/api/mesh/v1alpha1"
	konvoy_mesh_k8s "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/plugins/resources/k8s/native/api/v1alpha1"
	util_proto "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/util/proto"
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
		if templateName := proxy.Workload.Meta.GetAnnotations()[konvoy_mesh_k8s.ProxyTemplateAnnotation]; templateName != "" {
			k8sTemplate := &konvoy_mesh_k8s.ProxyTemplate{}
			if err := r.Client.Get(
				context.Background(),
				types.NamespacedName{Namespace: proxy.Workload.Meta.GetNamespace(), Name: templateName},
				k8sTemplate); err != nil {
				templateResolverLog.Error(err, "failed to resolve ProxyTemplate",
					"workloadNamespace", proxy.Workload.Meta.GetNamespace(),
					"workloadName", proxy.Workload.Meta.GetName(),
					"templateName", templateName)
			} else {
				templateResolverLog.V(1).Info("resolved ProxyTemplate",
					"workloadNamespace", proxy.Workload.Meta.GetNamespace(),
					"workloadName", proxy.Workload.Meta.GetName(),
					"templateName", templateName,
				)
				template := &konvoy_mesh.ProxyTemplate{}
				if err := util_proto.FromMap(k8sTemplate.Spec, template); err != nil {
					templateResolverLog.Error(err, "failed to unmarshal ProxyTemplate",
						"workloadNamespace", proxy.Workload.Meta.GetNamespace(),
						"workloadName", proxy.Workload.Meta.GetName(),
						"templateName", templateName)
				} else {
					return template
				}
			}
		}
	}
	return r.DefaultProxyTemplate
}
