package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	kube_apps "k8s.io/api/apps/v1"
	kube_core "k8s.io/api/core/v1"
	kube_apierrs "k8s.io/apimachinery/pkg/api/errors"
	kube_meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	kube_runtime "k8s.io/apimachinery/pkg/runtime"
	kube_schema "k8s.io/apimachinery/pkg/runtime/schema"
	kube_types "k8s.io/apimachinery/pkg/types"
	kube_ctrl "sigs.k8s.io/controller-runtime"
	kube_client "sigs.k8s.io/controller-runtime/pkg/client"
	kube_handler "sigs.k8s.io/controller-runtime/pkg/handler"
	kube_reconcile "sigs.k8s.io/controller-runtime/pkg/reconcile"
	kube_source "sigs.k8s.io/controller-runtime/pkg/source"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/manager"
	k8s_common "github.com/kumahq/kuma/pkg/plugins/common/k8s"
	mesh_k8s "github.com/kumahq/kuma/pkg/plugins/resources/k8s/native/api/v1alpha1"
	"github.com/kumahq/kuma/pkg/plugins/runtime/gateway/match"
	"github.com/kumahq/kuma/pkg/plugins/runtime/k8s/containers"
	ctrls_util "github.com/kumahq/kuma/pkg/plugins/runtime/k8s/controllers/util"
	"github.com/kumahq/kuma/pkg/plugins/runtime/k8s/metadata"
	k8s_util "github.com/kumahq/kuma/pkg/plugins/runtime/k8s/util"
)

// GatewayInstanceReconciler reconciles a MeshGatewayInstance object.
type GatewayInstanceReconciler struct {
	kube_client.Client
	Log logr.Logger

	Scheme          *kube_runtime.Scheme
	Converter       k8s_common.Converter
	ProxyFactory    *containers.DataplaneProxyFactory
	ResourceManager manager.ResourceManager
}

// Reconcile handles ensuring both a Service and a Deployment exist for an
// instance as well as setting the status.
func (r *GatewayInstanceReconciler) Reconcile(ctx context.Context, req kube_ctrl.Request) (kube_ctrl.Result, error) {
	gatewayInstance := &mesh_k8s.MeshGatewayInstance{}
	if err := r.Get(ctx, req.NamespacedName, gatewayInstance); err != nil {
		if kube_apierrs.IsNotFound(err) {
			return kube_ctrl.Result{}, nil
		}

		return kube_ctrl.Result{}, err
	}
	orig := gatewayInstance.DeepCopyObject().(kube_client.Object)

	svc, err := r.createOrUpdateService(ctx, gatewayInstance)
	if err != nil {
		return kube_ctrl.Result{}, errors.Wrap(err, "unable to reconcile Service for Gateway")
	}

	var deployment *kube_apps.Deployment
	if svc != nil {
		deployment, err = r.createOrUpdateDeployment(ctx, gatewayInstance)
		if err != nil {
			return kube_ctrl.Result{}, errors.Wrap(err, "unable to reconcile Deployment for Gateway")
		}
	}

	updateStatus(gatewayInstance, svc, deployment)

	if err := r.Client.Status().Patch(ctx, gatewayInstance, kube_client.MergeFrom(orig)); err != nil {
		if kube_apierrs.IsNotFound(err) {
			return kube_ctrl.Result{}, nil
		}
		return kube_ctrl.Result{}, errors.Wrap(err, "unable to patch MeshGatewayInstance status")
	}

	return kube_ctrl.Result{}, nil
}

func k8sSelector(name string) map[string]string {
	return map[string]string{"app": name}
}

// createOrUpdateService can either return an error, a created Service or
// neither if reconciliation shouldn't continue.
func (r *GatewayInstanceReconciler) createOrUpdateService(
	ctx context.Context,
	gatewayInstance *mesh_k8s.MeshGatewayInstance,
) (*kube_core.Service, error) {
	gatewayList := &core_mesh.MeshGatewayResourceList{}

	// XXX BUG: Needs to refer to specific mesh
	if err := r.ResourceManager.List(ctx, gatewayList); err != nil {
		return nil, err
	}
	gateway := match.Gateway(gatewayList, func(selector mesh_proto.TagSelector) bool {
		return selector.Matches(gatewayInstance.Spec.Tags)
	})

	obj, err := ctrls_util.ManageControlledObject(
		ctx, r.Client, gatewayInstance, &kube_core.ServiceList{},
		func(obj kube_client.Object) (kube_client.Object, error) {
			// If we don't have a gateway, we don't want our Service anymore
			if gateway == nil {
				return nil, nil
			}

			service := &kube_core.Service{
				ObjectMeta: kube_meta.ObjectMeta{
					Namespace:    gatewayInstance.Namespace,
					GenerateName: fmt.Sprintf("%s-", gatewayInstance.Name),
					Annotations: map[string]string{
						metadata.KumaGatewayAnnotation: metadata.AnnotationBuiltin,
					},
				},
			}
			if obj != nil {
				service = obj.(*kube_core.Service)
			}

			var ports []kube_core.ServicePort

			for _, listener := range gateway.Spec.GetConf().GetListeners() {
				servicePort := kube_core.ServicePort{
					Name:     strconv.Itoa(int(listener.Port)),
					Protocol: kube_core.ProtocolTCP,
					Port:     int32(listener.Port),
				}
				if gatewayInstance.Spec.ServiceType == kube_core.ServiceTypeNodePort {
					servicePort.NodePort = int32(listener.Port)
				}
				ports = append(ports, servicePort)
			}

			service.Spec.Selector = k8sSelector(gatewayInstance.Name)
			service.Spec.Ports = ports
			service.Spec.Type = gatewayInstance.Spec.ServiceType

			return service, nil
		},
	)
	if err != nil {
		return nil, err
	}

	if obj == nil {
		return nil, nil
	}

	return obj.(*kube_core.Service), nil
}

// createOrUpdateDeployment can either return an error, a created Deployment or
// neither if reconciliation shouldn't continue.
func (r *GatewayInstanceReconciler) createOrUpdateDeployment(
	ctx context.Context,
	gatewayInstance *mesh_k8s.MeshGatewayInstance,
) (*kube_apps.Deployment, error) {
	ns := kube_core.Namespace{}
	if err := r.Client.Get(ctx, kube_types.NamespacedName{Name: gatewayInstance.Namespace}, &ns); err != nil {
		return nil, errors.Wrap(err, "unable to get Namespace for MeshGatewayInstance")
	}

	obj, err := ctrls_util.ManageControlledObject(
		ctx, r.Client, gatewayInstance, &kube_apps.DeploymentList{},
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

			container, err := r.ProxyFactory.NewContainer(gatewayInstance, &ns)
			if err != nil {
				return nil, errors.Wrap(err, "unable to create gateway container")
			}

			if res := gatewayInstance.Spec.Resources; res != nil {
				container.Resources = *res
			}

			container.Name = k8s_util.KumaGatewayContainerName

			podSpec := kube_core.PodSpec{
				Containers: []kube_core.Container{container},
			}

			jsonTags, err := json.Marshal(gatewayInstance.Spec.Tags)
			if err != nil {
				return nil, errors.Wrap(err, "unable to marshal tags to JSON")
			}

			annotations := map[string]string{
				metadata.KumaGatewayAnnotation: metadata.AnnotationBuiltin,
				metadata.KumaTagsAnnotation:    string(jsonTags),
				metadata.KumaMeshAnnotation:    k8s_util.MeshOf(gatewayInstance, &ns),
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
		return nil, err
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

	switch svc.Spec.Type {
	case kube_core.ServiceTypeNodePort, kube_core.ServiceTypeClusterIP:
		// If we have any IP addresses assigned, the service is probably OK.
		for _, ip := range svc.Spec.ClusterIPs {
			if ip != kube_core.ClusterIPNone && ip != "" {
				return kube_meta.ConditionTrue, mesh_k8s.GatewayInstanceReady
			}
		}
	case kube_core.ServiceTypeLoadBalancer:
		if len(svc.Status.LoadBalancer.Ingress) > 0 {
			return kube_meta.ConditionTrue, mesh_k8s.GatewayInstanceReady
		}
	}

	return kube_meta.ConditionFalse, mesh_k8s.GatewayInstanceAddressNotReady
}

const noGateway = "No Gateway matched by tags"

func updateStatus(gatewayInstance *mesh_k8s.MeshGatewayInstance, svc *kube_core.Service, deployment *kube_apps.Deployment) {
	var status kube_meta.ConditionStatus
	var reason string
	var message string

	if svc == nil {
		status, reason, message = kube_meta.ConditionFalse, mesh_k8s.GatewayInstanceNoGatewayMatched, noGateway
	} else {
		status, reason = getCombinedReadiness(svc, deployment)
		gatewayInstance.Status.LoadBalancer = &svc.Status.LoadBalancer
	}

	readiness := kube_meta.Condition{
		Type: mesh_k8s.GatewayInstanceReady, Status: status, Reason: reason, Message: message, LastTransitionTime: kube_meta.Now(), ObservedGeneration: gatewayInstance.GetGeneration(),
	}

	gatewayInstance.Status.Conditions = []kube_meta.Condition{
		readiness,
	}
}

const serviceKey string = ".metadata.service"

// GatewayToInstanceMapper maps a Gateway object to MeshGatewayInstance objects by
// using the service tag to list GatewayInstances with a matching index.
// The index is set up on MeshGatewayInstance in SetupWithManager and holds the service
// tag from the MeshGatewayInstance tags.
func GatewayToInstanceMapper(l logr.Logger, client kube_client.Client) kube_handler.MapFunc {
	l = l.WithName("gateway-to-gateway-instance-mapper")

	return func(obj kube_client.Object) []kube_reconcile.Request {
		gateway := obj.(*mesh_k8s.MeshGateway)

		var serviceNames []string
		spec := gateway.GetSpec().(*mesh_proto.MeshGateway)
		for _, selector := range spec.GetSelectors() {
			if tagValue, ok := selector.Match[mesh_proto.ServiceTag]; ok {
				serviceNames = append(serviceNames, tagValue)
			}
		}

		var req []kube_reconcile.Request
		for _, serviceName := range serviceNames {
			instances := &mesh_k8s.MeshGatewayInstanceList{}
			if err := client.List(
				context.Background(), instances, kube_client.MatchingFields{serviceKey: serviceName},
			); err != nil {
				l.WithValues("gateway", obj.GetName()).Error(err, "failed to fetch GatewayInstances")
			}

			for _, instance := range instances.Items {
				req = append(req, kube_reconcile.Request{
					NamespacedName: kube_types.NamespacedName{Namespace: instance.Namespace, Name: instance.Name},
				})
			}
		}

		return req
	}
}

func (r *GatewayInstanceReconciler) SetupWithManager(mgr kube_ctrl.Manager) error {
	gatewayInstanceGVK := kube_schema.GroupVersionKind{
		Group:   mesh_k8s.GroupVersion.Group,
		Version: mesh_k8s.GroupVersion.Version,
		Kind:    reflect.TypeOf(mesh_k8s.MeshGatewayInstance{}).Name(),
	}

	if err := ctrls_util.IndexControllerOf(mgr, gatewayInstanceGVK, &kube_core.Service{}); err != nil {
		return err
	}

	if err := ctrls_util.IndexControllerOf(mgr, gatewayInstanceGVK, &kube_apps.Deployment{}); err != nil {
		return err
	}

	if err := mgr.GetFieldIndexer().IndexField(context.Background(), &mesh_k8s.MeshGatewayInstance{}, serviceKey, func(obj kube_client.Object) []string {
		instance := obj.(*mesh_k8s.MeshGatewayInstance)

		serviceName := instance.Spec.Tags[mesh_proto.ServiceTag]

		return []string{serviceName}
	}); err != nil {
		return err
	}

	return kube_ctrl.NewControllerManagedBy(mgr).
		For(&mesh_k8s.MeshGatewayInstance{}).
		Owns(&kube_core.Service{}).
		Owns(&kube_apps.Deployment{}).
		// On Update events our mapper function is called with the object both
		// before the event as well as the object after. In the case of
		// unbinding a Gateway from one Instance to another, we end up
		// reconciling both Instances.
		Watches(&kube_source.Kind{Type: &mesh_k8s.MeshGateway{}}, kube_handler.EnqueueRequestsFromMapFunc(GatewayToInstanceMapper(r.Log, mgr.GetClient()))).
		Complete(r)
}
