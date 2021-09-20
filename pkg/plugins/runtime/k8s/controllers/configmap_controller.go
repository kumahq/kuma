package controllers

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	kube_core "k8s.io/api/core/v1"
	kube_runtime "k8s.io/apimachinery/pkg/runtime"
	kube_types "k8s.io/apimachinery/pkg/types"
	kube_record "k8s.io/client-go/tools/record"
	kube_ctrl "sigs.k8s.io/controller-runtime"
	kube_client "sigs.k8s.io/controller-runtime/pkg/client"
	kube_handler "sigs.k8s.io/controller-runtime/pkg/handler"
	kube_reconile "sigs.k8s.io/controller-runtime/pkg/reconcile"
	kube_source "sigs.k8s.io/controller-runtime/pkg/source"

	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/manager"
	"github.com/kumahq/kuma/pkg/core/resources/store"
	"github.com/kumahq/kuma/pkg/dns"
	"github.com/kumahq/kuma/pkg/dns/vips"
	k8s_common "github.com/kumahq/kuma/pkg/plugins/common/k8s"
	mesh_k8s "github.com/kumahq/kuma/pkg/plugins/resources/k8s/native/api/v1alpha1"
	"github.com/kumahq/kuma/pkg/plugins/runtime/k8s/metadata"
)

// ConfigMapReconciler reconciles a ConfigMap object
type ConfigMapReconciler struct {
	kube_client.Client
	kube_record.EventRecorder
	Scheme            *kube_runtime.Scheme
	Log               logr.Logger
	ResourceManager   manager.ResourceManager
	ResourceConverter k8s_common.Converter
	VIPsAllocator     *dns.VIPsAllocator
	SystemNamespace   string
}

func (r *ConfigMapReconciler) Reconcile(req kube_ctrl.Request) (kube_ctrl.Result, error) {
	mesh, ok := vips.MeshFromConfigKey(req.Name)
	if !ok {
		return kube_ctrl.Result{}, nil
	}

	r.Log.V(1).Info("updating VIPs", "mesh", mesh)

	if err := r.VIPsAllocator.CreateOrUpdateVIPConfig(mesh); err != nil {
		if store.IsResourceConflict(err) {
			r.Log.V(1).Info("VIPs were updated in the other place. Retrying")
			return kube_ctrl.Result{Requeue: true}, nil
		}
		return kube_ctrl.Result{}, err
	}

	r.Log.V(1).Info("VIPs updated", "mesh", mesh)

	return kube_ctrl.Result{}, nil
}

func (r *ConfigMapReconciler) SetupWithManager(mgr kube_ctrl.Manager) error {
	return kube_ctrl.NewControllerManagedBy(mgr).
		For(&kube_core.ConfigMap{}).
		Watches(&kube_source.Kind{Type: &kube_core.Service{}}, &kube_handler.EnqueueRequestsFromMapFunc{
			ToRequests: &ServiceToConfigMapsMapper{
				Client:          mgr.GetClient(),
				Log:             r.Log.WithName("service-to-configmap-mapper"),
				SystemNamespace: r.SystemNamespace,
			},
		}).
		Watches(&kube_source.Kind{Type: &mesh_k8s.Dataplane{}}, &kube_handler.EnqueueRequestsFromMapFunc{
			ToRequests: &DataplaneToMeshMapper{
				Client:            mgr.GetClient(),
				Log:               r.Log.WithName("dataplane-to-configmap-mapper"),
				SystemNamespace:   r.SystemNamespace,
				ResourceConverter: r.ResourceConverter,
			},
		}).
		Watches(&kube_source.Kind{Type: &mesh_k8s.ZoneIngress{}}, &kube_handler.EnqueueRequestsFromMapFunc{
			ToRequests: &ZoneIngressToMeshMapper{
				Log:               r.Log.WithName("zone-ingress-to-configmap-mapper"),
				SystemNamespace:   r.SystemNamespace,
				ResourceConverter: r.ResourceConverter,
			},
		}).
		Watches(&kube_source.Kind{Type: &mesh_k8s.VirtualOutbound{}}, &kube_handler.EnqueueRequestsFromMapFunc{
			ToRequests: &VirtualOutboundToConfigMapsMapper{
				Client:          mgr.GetClient(),
				Log:             r.Log.WithName("virtualoutbound-to-configmap-mapper"),
				SystemNamespace: r.SystemNamespace,
			},
		}).
		Watches(&kube_source.Kind{Type: &mesh_k8s.ExternalService{}}, &kube_handler.EnqueueRequestsFromMapFunc{
			ToRequests: &ExternalServiceToConfigMapsMapper{
				Client:          mgr.GetClient(),
				Log:             r.Log.WithName("external-service-to-configmap-mapperr"),
				SystemNamespace: r.SystemNamespace,
			},
		}).
		Complete(r)
}

type ServiceToConfigMapsMapper struct {
	kube_client.Client
	Log             logr.Logger
	SystemNamespace string
}

func (m *ServiceToConfigMapsMapper) Map(obj kube_handler.MapObject) []kube_reconile.Request {
	cause, ok := obj.Object.(*kube_core.Service)
	if !ok {
		m.Log.WithValues("dataplane", obj.Meta).Error(errors.Errorf("wrong argument type: expected %T, got %T", cause, obj.Object), "wrong argument type")
		return nil
	}

	ctx := context.Background()
	svcName := fmt.Sprintf("%s/%s", cause.Namespace, cause.Name)
	// List Pods in the same namespace
	pods := &kube_core.PodList{}
	if err := m.Client.List(ctx, pods, kube_client.InNamespace(obj.Meta.GetNamespace())); err != nil {
		m.Log.WithValues("service", svcName).Error(err, "failed to fetch Dataplanes in namespace")
		return nil
	}

	meshSet := map[string]bool{}
	for _, pod := range pods.Items {
		if mesh, exist := metadata.Annotations(pod.Annotations).GetString(metadata.KumaMeshAnnotation); exist {
			meshSet[mesh] = true
		}
	}
	var req []kube_reconile.Request
	for mesh := range meshSet {
		req = append(req, kube_reconile.Request{
			NamespacedName: kube_types.NamespacedName{Namespace: m.SystemNamespace, Name: vips.ConfigKey(mesh)},
		})
	}

	return req
}

type DataplaneToMeshMapper struct {
	kube_client.Client
	Log               logr.Logger
	SystemNamespace   string
	ResourceConverter k8s_common.Converter
}

func (m *DataplaneToMeshMapper) Map(obj kube_handler.MapObject) []kube_reconile.Request {
	cause, ok := obj.Object.(*mesh_k8s.Dataplane)
	if !ok {
		m.Log.WithValues("dataplane", obj.Meta).Error(errors.Errorf("wrong argument type: expected %T, got %T", cause, obj.Object), "wrong argument type")
		return nil
	}

	dp := core_mesh.NewDataplaneResource()
	if err := m.ResourceConverter.ToCoreResource(cause, dp); err != nil {
		converterLog.Error(err, "failed to parse Dataplane", "dataplane", cause.Spec)
		return nil
	}

	// backwards compatibility
	if dp.Spec.IsIngress() {
		meshSet := map[string]bool{}
		for _, service := range dp.Spec.GetNetworking().GetIngress().GetAvailableServices() {
			meshSet[service.Mesh] = true
		}

		var requests []kube_reconile.Request
		for mesh := range meshSet {
			requests = append(requests, kube_reconile.Request{
				NamespacedName: kube_types.NamespacedName{Namespace: m.SystemNamespace, Name: vips.ConfigKey(mesh)},
			})
		}
		return requests
	}

	return []kube_reconile.Request{{
		NamespacedName: kube_types.NamespacedName{Namespace: m.SystemNamespace, Name: vips.ConfigKey(cause.Mesh)},
	}}
}

type ZoneIngressToMeshMapper struct {
	Log               logr.Logger
	ResourceConverter k8s_common.Converter
	SystemNamespace   string
}

func (m *ZoneIngressToMeshMapper) Map(obj kube_handler.MapObject) []kube_reconile.Request {
	cause, ok := obj.Object.(*mesh_k8s.ZoneIngress)
	if !ok {
		m.Log.WithValues("zoneIngress", obj.Meta).Error(errors.Errorf("wrong argument type: expected %T, got %T", cause, obj.Object), "wrong argument type")
		return nil
	}
	zoneIngress := core_mesh.NewZoneIngressResource()
	if err := m.ResourceConverter.ToCoreResource(cause, zoneIngress); err != nil {
		converterLog.Error(err, "failed to parse ZoneIngress", "zoneIngress", cause.Spec)
		return nil
	}

	meshSet := map[string]bool{}
	for _, service := range zoneIngress.Spec.GetAvailableServices() {
		meshSet[service.Mesh] = true
	}

	var requests []kube_reconile.Request
	for mesh := range meshSet {
		requests = append(requests, kube_reconile.Request{
			NamespacedName: kube_types.NamespacedName{Namespace: m.SystemNamespace, Name: vips.ConfigKey(mesh)},
		})
	}
	return requests
}

type ExternalServiceToConfigMapsMapper struct {
	kube_client.Client
	Log             logr.Logger
	SystemNamespace string
}

func (m *ExternalServiceToConfigMapsMapper) Map(obj kube_handler.MapObject) []kube_reconile.Request {
	cause, ok := obj.Object.(*mesh_k8s.ExternalService)
	if !ok {
		m.Log.WithValues("externalService", obj.Meta).Error(errors.Errorf("wrong argument type: expected %T, got %T", cause, obj.Object), "wrong argument type")
		return nil
	}

	return []kube_reconile.Request{{
		NamespacedName: kube_types.NamespacedName{Namespace: m.SystemNamespace, Name: vips.ConfigKey(cause.Mesh)},
	}}
}

type VirtualOutboundToConfigMapsMapper struct {
	kube_client.Client
	Log             logr.Logger
	SystemNamespace string
}

func (m *VirtualOutboundToConfigMapsMapper) Map(obj kube_handler.MapObject) []kube_reconile.Request {
	cause, ok := obj.Object.(*mesh_k8s.VirtualOutbound)
	if !ok {
		m.Log.WithValues("virtualOutbound", obj.Meta).Error(errors.Errorf("wrong argument type: expected %T, got %T", cause, obj.Object), "wrong argument type")
		return nil
	}

	return []kube_reconile.Request{{
		NamespacedName: kube_types.NamespacedName{Namespace: m.SystemNamespace, Name: vips.ConfigKey(cause.Mesh)},
	}}
}
