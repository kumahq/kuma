package controllers

import (
	"context"
	"fmt"
	"reflect"
	"strconv"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	kube_apps "k8s.io/api/apps/v1"
	kube_core "k8s.io/api/core/v1"
	kube_apierrs "k8s.io/apimachinery/pkg/api/errors"
	kube_apimeta "k8s.io/apimachinery/pkg/api/meta"
	kube_meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	kube_runtime "k8s.io/apimachinery/pkg/runtime"
	kube_schema "k8s.io/apimachinery/pkg/runtime/schema"
	kube_types "k8s.io/apimachinery/pkg/types"
	kube_ctrl "sigs.k8s.io/controller-runtime"
	kube_client "sigs.k8s.io/controller-runtime/pkg/client"
	kube_controllerutil "sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/resources/manager"
	k8s_common "github.com/kumahq/kuma/pkg/plugins/common/k8s"
	mesh_k8s "github.com/kumahq/kuma/pkg/plugins/resources/k8s/native/api/v1alpha1"
	"github.com/kumahq/kuma/pkg/plugins/runtime/gateway/match"
	"github.com/kumahq/kuma/pkg/plugins/runtime/k8s/containers"
	"github.com/kumahq/kuma/pkg/plugins/runtime/k8s/metadata"
	util_k8s "github.com/kumahq/kuma/pkg/plugins/runtime/k8s/util"
)

// GatewayInstanceReconciler reconciles a GatewayInstance object.
type GatewayInstanceReconciler struct {
	kube_client.Client
	Log logr.Logger

	Scheme          *kube_runtime.Scheme
	Converter       k8s_common.Converter
	ProxyFactory    containers.DataplaneProxyFactory
	ResourceManager manager.ResourceManager
}

// Reconcile handles ensuring both a Service and a Deployment exist for an
// instance as well as setting the status.
func (r *GatewayInstanceReconciler) Reconcile(ctx context.Context, req kube_ctrl.Request) (kube_ctrl.Result, error) {
	gatewayInstance := &mesh_k8s.GatewayInstance{}
	if err := r.Get(ctx, req.NamespacedName, gatewayInstance); err != nil {
		if kube_apierrs.IsNotFound(err) {
			return kube_ctrl.Result{}, nil
		}

		return kube_ctrl.Result{}, err
	}

	deployment, err := r.createOrUpdateDeployment(ctx, gatewayInstance)
	if err != nil {
		return kube_ctrl.Result{}, errors.Wrap(err, "unable to create Deployment for Gateway")
	}

	svc, err := r.createOrUpdateService(ctx, gatewayInstance)
	if err != nil {
		return kube_ctrl.Result{}, errors.Wrap(err, "unable to create Service for Gateway")
	}

	updateStatus(gatewayInstance, svc, deployment)

	if err := r.Client.Status().Update(ctx, gatewayInstance); err != nil {
		return kube_ctrl.Result{}, errors.Wrap(err, "unable to update GatewayInstance status")
	}

	return kube_ctrl.Result{}, nil
}

func k8sSelector(name string) map[string]string {
	return map[string]string{"app": name}
}

const controllerKey string = ".metadata.controller"

func indexController(mgr kube_ctrl.Manager, ownerGVK kube_schema.GroupVersionKind, objType kube_client.Object) error {
	ownerKind := ownerGVK.Kind
	ownerGV := ownerGVK.GroupVersion().String()

	return mgr.GetFieldIndexer().IndexField(context.Background(), objType, controllerKey, func(rawObj kube_client.Object) []string {
		owner := kube_meta.GetControllerOf(rawObj)
		if owner == nil || owner.APIVersion != ownerGV || owner.Kind != ownerKind {
			return nil
		}

		return []string{owner.Name}
	})
}

// createOrUpdateControlled either creates an object to be controlled by the
// given object or updates the existing one.
func (r *GatewayInstanceReconciler) createOrUpdateControlled(
	ctx context.Context, owner kube_meta.Object, objectList kube_client.ObjectList, mutate func(kube_client.Object) (kube_client.Object, error),
) (kube_client.Object, error) {
	if err := r.Client.List(
		ctx, objectList, kube_client.InNamespace(owner.GetNamespace()), kube_client.MatchingFields{controllerKey: owner.GetName()},
	); err != nil {
		return nil, errors.Wrap(err, "unable to list objects")
	}

	items, err := kube_apimeta.ExtractList(objectList)
	if err != nil {
		return nil, errors.Wrap(err, "unable to extract list of runtime.Objects")
	}

	switch len(items) {
	case 0:
		obj, err := mutate(nil)
		if err != nil {
			return nil, errors.Wrap(err, "couldn't mutate object for creation")
		}

		if err := kube_controllerutil.SetControllerReference(owner, obj, r.Scheme); err != nil {
			return nil, errors.Wrap(err, "unable to set object's controller reference")
		}

		if err := r.Client.Create(ctx, obj); err != nil {
			return nil, errors.Wrap(err, "couldn't create object")
		}
		return obj, nil
	case 1:
		item, ok := items[0].(kube_client.Object)
		if !ok {
			return nil, fmt.Errorf("expected runtime.Object to be client.Object")
		}

		obj, err := mutate(item)
		if err != nil {
			return nil, errors.Wrap(err, "couldn't mutate object for update")
		}

		if err := kube_controllerutil.SetControllerReference(owner, obj, r.Scheme); err != nil {
			return nil, errors.Wrap(err, "unable to set object's controller reference")
		}

		if err := r.Client.Update(ctx, obj); err != nil {
			return nil, errors.Wrap(err, "couldn't update object")
		}
		return obj, nil
	default:
		return nil, fmt.Errorf("expected a maximum of one item controlled by %s/%s", owner.GetNamespace(), owner.GetName())
	}
}

func (r *GatewayInstanceReconciler) createOrUpdateService(
	ctx context.Context,
	gatewayInstance *mesh_k8s.GatewayInstance,
) (*kube_core.Service, error) {
	gateway := match.Gateway(r.ResourceManager, func(selector mesh_proto.TagSelector) bool {
		return selector.Matches(gatewayInstance.Spec.Tags)
	})

	if gateway == nil {
		return nil, fmt.Errorf("no matching Gateway")
	}

	obj, err := r.createOrUpdateControlled(
		ctx, gatewayInstance, &kube_core.ServiceList{},
		func(obj kube_client.Object) (kube_client.Object, error) {
			service := &kube_core.Service{
				ObjectMeta: kube_meta.ObjectMeta{
					Namespace:    gatewayInstance.Namespace,
					GenerateName: fmt.Sprintf("%s-", gatewayInstance.Name),
				},
			}
			if obj != nil {
				service = obj.(*kube_core.Service)
			}

			var ports []kube_core.ServicePort

			for _, listener := range gateway.Spec.GetConf().GetListeners() {
				ports = append(ports, kube_core.ServicePort{
					Name:     strconv.Itoa(int(listener.Port)),
					Protocol: kube_core.ProtocolTCP,
					Port:     int32(listener.Port),
				})
			}

			service.Spec = kube_core.ServiceSpec{
				Selector: k8sSelector(gatewayInstance.Name),
				Ports:    ports,
				Type:     gatewayInstance.Spec.ServiceType,
			}

			return service, nil
		},
	)
	if err != nil {
		return nil, errors.Wrap(err, "unable to create or update Service for GatewayInstance")
	}

	return obj.(*kube_core.Service), nil
}

func (r *GatewayInstanceReconciler) createOrUpdateDeployment(
	ctx context.Context,
	gatewayInstance *mesh_k8s.GatewayInstance,
) (*kube_apps.Deployment, error) {
	ns := kube_core.Namespace{}
	if err := r.Client.Get(ctx, kube_types.NamespacedName{Name: gatewayInstance.Namespace}, &ns); err != nil {
		return nil, errors.Wrap(err, "unable to get Namespace for GatewayInstance")
	}

	obj, err := r.createOrUpdateControlled(
		ctx, gatewayInstance, &kube_apps.DeploymentList{},
		func(obj kube_client.Object) (kube_client.Object, error) {
			deployment := &kube_apps.Deployment{
				ObjectMeta: kube_meta.ObjectMeta{
					Namespace:    gatewayInstance.Namespace,
					GenerateName: fmt.Sprintf("%s-", gatewayInstance.Name),
				},
			}
			if obj != nil {
				deployment = obj.(*kube_apps.Deployment)
			}

			container, err := r.ProxyFactory.NewContainer(gatewayInstance.Annotations, &ns)
			if err != nil {
				return nil, errors.Wrap(err, "unable to create gateway container")
			}

			if res := gatewayInstance.Spec.Resources; res != nil {
				container.Resources = *res
			}

			container.Name = util_k8s.KumaGatewayContainerName

			podSpec := kube_core.PodSpec{
				Containers: []kube_core.Container{container},
			}

			annotations := map[string]string{
				metadata.KumaGatewayAnnotation: metadata.AnnotationBuiltin,
				mesh_proto.ServiceTag:          gatewayInstance.Spec.Tags[mesh_proto.ServiceTag],
				metadata.KumaMeshAnnotation:    util_k8s.MeshFor(gatewayInstance),
			}

			labels := k8sSelector(gatewayInstance.Name)
			labels[metadata.KumaSidecarInjectionAnnotation] = metadata.AnnotationDisabled

			deployment.Spec.Replicas = &gatewayInstance.Spec.Replicas
			deployment.Spec.Selector = &kube_meta.LabelSelector{
				MatchLabels: k8sSelector(gatewayInstance.Name),
			}
			deployment.Spec.Template = kube_core.PodTemplateSpec{
				ObjectMeta: kube_meta.ObjectMeta{
					Labels:      labels,
					Annotations: annotations,
				},
				Spec: podSpec,
			}

			return deployment, nil
		},
	)
	if err != nil {
		return nil, errors.Wrap(err, "unable to create or update Service for GatewayInstance")
	}

	return obj.(*kube_apps.Deployment), nil
}

func getCombinedReadiness(svc *kube_core.Service, deployment *kube_apps.Deployment) (kube_meta.ConditionStatus, string) {
	for _, c := range deployment.Status.Conditions {
		if c.Type != kube_apps.DeploymentAvailable {
			continue
		}
		switch c.Status {
		case kube_core.ConditionFalse:
			return kube_meta.ConditionFalse, mesh_k8s.GatewayInstanceDeploymentNotAvailable
		case kube_core.ConditionUnknown:
			return kube_meta.ConditionUnknown, mesh_k8s.GatewayInstanceDeploymentNotAvailable
		case kube_core.ConditionTrue:
			break
		}
	}

	if len(svc.Status.LoadBalancer.Ingress) > 0 {
		return kube_meta.ConditionTrue, mesh_k8s.GatewayInstanceReady
	} else {
		return kube_meta.ConditionFalse, mesh_k8s.GatewayInstanceAddressNotReady
	}
}

func updateStatus(gatewayInstance *mesh_k8s.GatewayInstance, svc *kube_core.Service, deployment *kube_apps.Deployment) {
	status, reason := getCombinedReadiness(svc, deployment)

	readiness := kube_meta.Condition{
		Type: mesh_k8s.GatewayInstanceReady, Status: status, Reason: reason, LastTransitionTime: kube_meta.Now(), ObservedGeneration: gatewayInstance.GetGeneration(),
	}

	gatewayInstance.Status.LoadBalancer = &svc.Status.LoadBalancer

	gatewayInstance.Status.Conditions = []kube_meta.Condition{
		readiness,
	}
}

func (r *GatewayInstanceReconciler) SetupWithManager(mgr kube_ctrl.Manager) error {
	gatewayInstanceGVK := kube_schema.GroupVersionKind{
		Group:   mesh_k8s.GroupVersion.Group,
		Version: mesh_k8s.GroupVersion.Version,
		Kind:    reflect.TypeOf(mesh_k8s.GatewayInstance{}).Name(),
	}

	if err := indexController(mgr, gatewayInstanceGVK, &kube_core.Service{}); err != nil {
		return err
	}

	if err := indexController(mgr, gatewayInstanceGVK, &kube_apps.Deployment{}); err != nil {
		return err
	}

	return kube_ctrl.NewControllerManagedBy(mgr).
		For(&mesh_k8s.GatewayInstance{}).
		Owns(&kube_core.Service{}).
		Owns(&kube_apps.Deployment{}).
		Complete(r)
}
