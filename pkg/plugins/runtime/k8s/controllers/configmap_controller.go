package controllers

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	"go.uber.org/multierr"
	kube_core "k8s.io/api/core/v1"
	kube_runtime "k8s.io/apimachinery/pkg/runtime"
	kube_types "k8s.io/apimachinery/pkg/types"
	kube_record "k8s.io/client-go/tools/record"
	kube_ctrl "sigs.k8s.io/controller-runtime"
	kube_client "sigs.k8s.io/controller-runtime/pkg/client"
	kube_handler "sigs.k8s.io/controller-runtime/pkg/handler"
	kube_reconile "sigs.k8s.io/controller-runtime/pkg/reconcile"
	kube_source "sigs.k8s.io/controller-runtime/pkg/source"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/manager"
	"github.com/kumahq/kuma/pkg/core/resources/store"
	"github.com/kumahq/kuma/pkg/dns"
	"github.com/kumahq/kuma/pkg/dns/persistence"
	mesh_k8s "github.com/kumahq/kuma/pkg/plugins/resources/k8s/native/api/v1alpha1"
	"github.com/kumahq/kuma/pkg/plugins/runtime/k8s/metadata"
)

// ConfigMapReconciler reconciles a ConfigMap object
type ConfigMapReconciler struct {
	kube_client.Client
	kube_record.EventRecorder
	Scheme          *kube_runtime.Scheme
	Log             logr.Logger
	ResourceManager manager.ResourceManager
	IPAM            dns.IPAM
	Persistence     persistence.MeshedWriter
	Resolver        dns.DNSResolver
}

func (r *ConfigMapReconciler) Reconcile(req kube_ctrl.Request) (kube_ctrl.Result, error) {
	mesh, ok := persistence.MeshedConfigKey(req.Name)
	if !ok {
		return kube_ctrl.Result{}, nil
	}
	serviceSet := map[string]bool{}
	dataplanes := core_mesh.DataplaneResourceList{}
	err := r.ResourceManager.List(context.Background(), &dataplanes, store.ListByMesh(mesh))
	if err != nil {
		return kube_ctrl.Result{}, err
	}
	for _, dp := range dataplanes.Items {
		if dp.Spec.IsIngress() {
			for _, service := range dp.Spec.Networking.Ingress.AvailableServices {
				serviceSet[service.Tags[mesh_proto.ServiceTag]] = true
			}
		} else {
			for _, inbound := range dp.Spec.Networking.Inbound {
				serviceSet[inbound.GetService()] = true
			}
		}
	}

	externalServices := core_mesh.ExternalServiceResourceList{}
	err = r.ResourceManager.List(context.Background(), &externalServices, store.ListByMesh(mesh))
	if err != nil {
		return kube_ctrl.Result{}, err
	}

	for _, es := range externalServices.Items {
		serviceSet[es.Spec.GetService()] = true
	}

	vips, err := r.Persistence.GetByMesh(mesh)
	if err != nil {
		return kube_ctrl.Result{}, err
	}
	change := false

	var errs error
	// ensure all services have entries in the domain
	for service := range serviceSet {
		_, found := vips[service]
		if !found {
			ip, err := r.IPAM.AllocateIP()
			if err != nil {
				errs = multierr.Append(errs, errors.Wrapf(err, "unable to allocate an ip for service %s", service))
			} else {
				vips[service] = ip
				change = true
			}
		}
	}

	// ensure all entries in the domain are present in the service list, and delete them otherwise
	for service := range vips {
		_, found := serviceSet[service]
		if !found {
			ip := vips[service]
			change = true
			_ = r.IPAM.FreeIP(ip)
			delete(vips, service)
		}
	}

	if change {
		if err := r.Persistence.Set(mesh, vips); err != nil {
			return kube_ctrl.Result{}, multierr.Append(errs, err)
		}
		r.Resolver.SetVIPs(vips)
	}

	return kube_ctrl.Result{}, nil
}

func (r *ConfigMapReconciler) SetupWithManager(mgr kube_ctrl.Manager) error {
	for _, addToScheme := range []func(*kube_runtime.Scheme) error{kube_core.AddToScheme, mesh_k8s.AddToScheme} {
		if err := addToScheme(mgr.GetScheme()); err != nil {
			return err
		}
	}
	return kube_ctrl.NewControllerManagedBy(mgr).
		For(&kube_core.ConfigMap{}).
		Watches(&kube_source.Kind{Type: &kube_core.Service{}}, &kube_handler.EnqueueRequestsFromMapFunc{
			ToRequests: &ServiceToConfigMapsMapper{Client: mgr.GetClient(), Log: r.Log.WithName("service-to-configmap-mapper")},
		}).
		Watches(&kube_source.Kind{Type: &mesh_k8s.Dataplane{}}, &kube_handler.EnqueueRequestsFromMapFunc{
			ToRequests: &DataplaneToMeshMapper{Client: mgr.GetClient(), Log: r.Log.WithName("dataplane-to-configmap-mapper")},
		}).
		Watches(&kube_source.Kind{Type: &mesh_k8s.ExternalService{}}, &kube_handler.EnqueueRequestsFromMapFunc{
			ToRequests: &ExternalServiceToConfigMapsMapper{Client: mgr.GetClient(), Log: r.Log.WithName("external-service-to-configmap-mapperr")},
		}).
		Complete(r)
}

type ServiceToConfigMapsMapper struct {
	kube_client.Client
	Log logr.Logger
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
		meshSet[pod.Annotations[metadata.KumaMeshAnnotation]] = true
	}
	var req []kube_reconile.Request
	for mesh := range meshSet {
		req = append(req, kube_reconile.Request{
			NamespacedName: kube_types.NamespacedName{Namespace: "kuma-system", Name: fmt.Sprintf("kuma-%s-dns-vips", mesh)},
		})
	}

	return req
}

type DataplaneToMeshMapper struct {
	kube_client.Client
	Log logr.Logger
}

func (m *DataplaneToMeshMapper) Map(obj kube_handler.MapObject) []kube_reconile.Request {
	cause, ok := obj.Object.(*mesh_k8s.Dataplane)
	if !ok {
		m.Log.WithValues("dataplane", obj.Meta).Error(errors.Errorf("wrong argument type: expected %T, got %T", cause, obj.Object), "wrong argument type")
		return nil
	}

	return []kube_reconile.Request{{
		NamespacedName: kube_types.NamespacedName{Namespace: "kuma-system", Name: fmt.Sprintf("kuma-%s-dns-vips", cause.Mesh)},
	}}
}

type ExternalServiceToConfigMapsMapper struct {
	kube_client.Client
	Log logr.Logger
}

func (m *ExternalServiceToConfigMapsMapper) Map(obj kube_handler.MapObject) []kube_reconile.Request {
	cause, ok := obj.Object.(*mesh_k8s.ExternalService)
	if !ok {
		m.Log.WithValues("externalService", obj.Meta).Error(errors.Errorf("wrong argument type: expected %T, got %T", cause, obj.Object), "wrong argument type")
		return nil
	}

	return []kube_reconile.Request{{
		NamespacedName: kube_types.NamespacedName{Namespace: "kuma-system", Name: fmt.Sprintf("kuma-%s-dns-vips", cause.Mesh)},
	}}
}