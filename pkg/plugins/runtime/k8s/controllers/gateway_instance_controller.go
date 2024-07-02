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
	kube_apimeta "k8s.io/apimachinery/pkg/api/meta"
	kube_meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	kube_runtime "k8s.io/apimachinery/pkg/runtime"
	kube_schema "k8s.io/apimachinery/pkg/runtime/schema"
	kube_types "k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	kube_version "k8s.io/apimachinery/pkg/version"
	kube_ctrl "sigs.k8s.io/controller-runtime"
	kube_client "sigs.k8s.io/controller-runtime/pkg/client"
	kube_handler "sigs.k8s.io/controller-runtime/pkg/handler"
	kube_reconcile "sigs.k8s.io/controller-runtime/pkg/reconcile"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/manager"
	"github.com/kumahq/kuma/pkg/core/resources/store"
	k8s_common "github.com/kumahq/kuma/pkg/plugins/common/k8s"
	mesh_k8s "github.com/kumahq/kuma/pkg/plugins/resources/k8s/native/api/v1alpha1"
	"github.com/kumahq/kuma/pkg/plugins/runtime/k8s/containers"
	ctrls_util "github.com/kumahq/kuma/pkg/plugins/runtime/k8s/controllers/util"
	"github.com/kumahq/kuma/pkg/plugins/runtime/k8s/metadata"
	k8s_util "github.com/kumahq/kuma/pkg/plugins/runtime/k8s/util"
	"github.com/kumahq/kuma/pkg/util/pointer"
	xds_topology "github.com/kumahq/kuma/pkg/xds/topology"
)

// GatewayInstanceReconciler reconciles a MeshGatewayInstance object.
type GatewayInstanceReconciler struct {
	K8sVersion *kube_version.Info
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
	r.Log.V(1).Info("reconcile", "req", req)
	gatewayInstance := &mesh_k8s.MeshGatewayInstance{}
	if err := r.Get(ctx, req.NamespacedName, gatewayInstance); err != nil {
		if kube_apierrs.IsNotFound(err) {
			return kube_ctrl.Result{}, nil
		}

		return kube_ctrl.Result{}, err
	}

	ns := kube_core.Namespace{}
	if err := r.Client.Get(ctx, kube_types.NamespacedName{Name: gatewayInstance.Namespace}, &ns); err != nil {
		return kube_ctrl.Result{}, errors.Wrap(err, "unable to get Namespace of MeshGatewayInstance")
	}

	mesh := k8s_util.MeshOfByLabelOrAnnotation(r.Log, gatewayInstance, &ns)

	orig := gatewayInstance.DeepCopyObject().(kube_client.Object)
	svc, gateway, err := r.createOrUpdateService(ctx, mesh, gatewayInstance)
	if err != nil {
		return kube_ctrl.Result{}, errors.Wrap(err, "unable to reconcile Service for Gateway")
	}

	var deployment *kube_apps.Deployment
	if svc != nil {
		deployment, err = r.createOrUpdateDeployment(ctx, mesh, gatewayInstance)
		if err != nil {
			return kube_ctrl.Result{}, errors.Wrap(err, "unable to reconcile Deployment for Gateway")
		}
	}

	updateStatus(gatewayInstance, gateway, mesh, svc, deployment)

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
	mesh string,
	gatewayInstance *mesh_k8s.MeshGatewayInstance,
) (*kube_core.Service, *core_mesh.MeshGatewayResource, error) {
	gatewayList := &core_mesh.MeshGatewayResourceList{}
	if err := r.ResourceManager.List(ctx, gatewayList, store.ListByMesh(mesh)); err != nil {
		return nil, nil, err
	}
	gateway := xds_topology.SelectGateway(gatewayList.Items, func(selector mesh_proto.TagSelector) bool {
		return selector.Matches(gatewayInstance.Spec.Tags)
	})

	obj, err := ctrls_util.ManageControlledObject(
		ctx, r.Client, gatewayInstance, &kube_core.ServiceList{},
		func(obj kube_client.Object) (kube_client.Object, error) {
			// If we don't have a gateway, we don't change anything. If the Service was already created, we keep it.
			// If there is no Service, we don't create one. We don't want to break the traffic if MeshGateway is absent
			// for a short period of time (i.e. due to renaming).
			if gateway == nil {
				return obj, nil
			}

			svcAnnotations := map[string]string{metadata.KumaGatewayAnnotation: metadata.AnnotationBuiltin}
			svcLabels := map[string]string{}

			if obj != nil {
				for k, v := range obj.GetAnnotations() {
					svcAnnotations[k] = v
				}
				for k, v := range obj.GetLabels() {
					svcLabels[k] = v
				}
			}

			for k, v := range gatewayInstance.Spec.ServiceTemplate.Metadata.Annotations {
				svcAnnotations[k] = v
			}

			for k, v := range gatewayInstance.Spec.ServiceTemplate.Metadata.Labels {
				svcLabels[k] = v
			}

			service := &kube_core.Service{
				ObjectMeta: kube_meta.ObjectMeta{
					Namespace: gatewayInstance.Namespace,
					Name:      gatewayInstance.Name,
				},
			}
			if obj != nil {
				service = obj.(*kube_core.Service)
			}

			service.Annotations = svcAnnotations
			service.Labels = svcLabels

			var ports []kube_core.ServicePort
			seenPorts := map[uint32]struct{}{}

			for i, listener := range gateway.Spec.GetConf().GetListeners() {
				if _, sawPort := seenPorts[listener.Port]; sawPort {
					// There can be multiple listeners with the same port
					// if they have different hostnames
					continue
				}
				seenPorts[listener.Port] = struct{}{}

				servicePort := kube_core.ServicePort{}
				if len(service.Spec.Ports) > i {
					// Reuse existing port to avoid mutating the object.
					servicePort = service.Spec.Ports[i]
				}

				servicePort.Name = strconv.Itoa(int(listener.Port))
				servicePort.Protocol = kube_core.ProtocolTCP
				servicePort.Port = int32(listener.Port)
				servicePort.TargetPort = intstr.FromInt(int(servicePort.Port))
				if gatewayInstance.Spec.ServiceType == kube_core.ServiceTypeNodePort {
					servicePort.NodePort = int32(listener.Port)
				}
				ports = append(ports, servicePort)
			}

			service.Spec.Selector = k8sSelector(gatewayInstance.Name)
			service.Spec.Ports = ports
			service.Spec.Type = gatewayInstance.Spec.ServiceType

			if ip := gatewayInstance.Spec.ServiceTemplate.Spec.LoadBalancerIP; ip != "" {
				service.Spec.LoadBalancerIP = ip
			}

			return service, nil
		},
	)
	if err != nil {
		return nil, gateway, err
	}

	if obj == nil {
		return nil, gateway, nil
	}

	return obj.(*kube_core.Service), gateway, nil
}

// createOrUpdateDeployment can either return an error, a created Deployment or
// neither if reconciliation shouldn't continue.
func (r *GatewayInstanceReconciler) createOrUpdateDeployment(
	ctx context.Context,
	mesh string,
	gatewayInstance *mesh_k8s.MeshGatewayInstance,
) (*kube_apps.Deployment, error) {
	obj, err := ctrls_util.ManageControlledObject(
		ctx, r.Client, gatewayInstance, &kube_apps.DeploymentList{},
		func(obj kube_client.Object) (kube_client.Object, error) {
			deployment := &kube_apps.Deployment{
				ObjectMeta: kube_meta.ObjectMeta{
					Namespace: gatewayInstance.Namespace,
					Name:      gatewayInstance.Name,
				},
			}
			if obj != nil {
				deployment = obj.(*kube_apps.Deployment)
			}

			container, err := r.ProxyFactory.NewContainer(gatewayInstance, mesh)
			if err != nil {
				return nil, errors.Wrap(err, "unable to create gateway container")
			}

			if res := gatewayInstance.Spec.Resources; res != nil {
				container.Resources = *res
			}

			container.Name = k8s_util.KumaGatewayContainerName
			container.Env = append(container.Env,
				kube_core.EnvVar{
					Name: "KUMA_DATAPLANE_RESOURCES_MAX_MEMORY_BYTES",
					ValueFrom: &kube_core.EnvVarSource{
						ResourceFieldRef: &kube_core.ResourceFieldSelector{
							ContainerName: container.Name,
							Resource:      "limits.memory",
						},
					},
				},
			)

			container.SecurityContext.AllowPrivilegeEscalation = pointer.To(false)

			securityContext := &kube_core.PodSecurityContext{
				Sysctls: []kube_core.Sysctl{{
					Name:  "net.ipv4.ip_unprivileged_port_start",
					Value: "0",
				}},
			}

			if fsGroup := gatewayInstance.Spec.PodTemplate.Spec.PodSecurityContext.FSGroup; fsGroup != nil {
				securityContext.FSGroup = fsGroup
			}

			container.SecurityContext.ReadOnlyRootFilesystem = gatewayInstance.Spec.PodTemplate.Spec.Container.SecurityContext.ReadOnlyRootFilesystem

			container.VolumeMounts = []kube_core.VolumeMount{
				{
					Name:      "tmp",
					MountPath: "/tmp",
				},
			}

			podSpec := kube_core.PodSpec{
				SecurityContext: securityContext,
				Containers:      []kube_core.Container{container},
			}
			podSpec.ServiceAccountName = gatewayInstance.Spec.PodTemplate.Spec.ServiceAccountName

			podSpec.Volumes = []kube_core.Volume{
				{
					Name: "tmp",
					VolumeSource: kube_core.VolumeSource{
						EmptyDir: &kube_core.EmptyDirVolumeSource{},
					},
				},
			}

			jsonTags, err := json.Marshal(gatewayInstance.Spec.Tags)
			if err != nil {
				return nil, errors.Wrap(err, "unable to marshal tags to JSON")
			}

			podAnnotations := map[string]string{
				metadata.KumaGatewayAnnotation: metadata.AnnotationBuiltin,
				metadata.KumaTagsAnnotation:    string(jsonTags),
				metadata.KumaMeshAnnotation:    mesh,
			}

			if obj != nil {
				for k, v := range obj.(*kube_apps.Deployment).Spec.Template.GetAnnotations() {
					podAnnotations[k] = v
				}
			}

			for k, v := range gatewayInstance.Spec.PodTemplate.Metadata.Annotations {
				podAnnotations[k] = v
			}

			podLabels := k8sSelector(gatewayInstance.Name)
			podLabels[metadata.KumaSidecarInjectionAnnotation] = metadata.AnnotationDisabled

			for k, v := range gatewayInstance.Spec.PodTemplate.Metadata.Labels {
				podLabels[k] = v
			}

			deployment.Spec.Replicas = &gatewayInstance.Spec.Replicas
			deployment.Spec.Selector = &kube_meta.LabelSelector{
				MatchLabels: k8sSelector(gatewayInstance.Name),
			}
			deployment.Spec.Template = kube_core.PodTemplateSpec{
				ObjectMeta: kube_meta.ObjectMeta{
					Labels:      podLabels,
					Annotations: podAnnotations,
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

func updateStatus(
	gatewayInstance *mesh_k8s.MeshGatewayInstance,
	gateway *core_mesh.MeshGatewayResource,
	mesh string,
	svc *kube_core.Service,
	deployment *kube_apps.Deployment,
) {
	var status kube_meta.ConditionStatus
	var reason string
	var message string

	if gateway == nil {
		status, reason, message = kube_meta.ConditionFalse, mesh_k8s.GatewayInstanceNoGatewayMatched, fmt.Sprintf("No Gateway matched by tags in mesh: '%s'", mesh)
	} else {
		status, reason = getCombinedReadiness(svc, deployment)
		gatewayInstance.Status.LoadBalancer = &svc.Status.LoadBalancer
	}

	readiness := kube_meta.Condition{
		Type:               mesh_k8s.GatewayInstanceReady,
		Status:             status,
		Reason:             reason,
		Message:            message,
		ObservedGeneration: gatewayInstance.GetGeneration(),
	}

	kube_apimeta.SetStatusCondition(
		&gatewayInstance.Status.Conditions,
		readiness,
	)
}

const serviceKey string = ".metadata.service"

// GatewayToInstanceMapper maps a Gateway object to MeshGatewayInstance objects by
// using the service tag to list GatewayInstances with a matching index.
// The index is set up on MeshGatewayInstance in SetupWithManager and holds the service
// tag from the MeshGatewayInstance tags.
func GatewayToInstanceMapper(l logr.Logger, client kube_client.Client) kube_handler.MapFunc {
	l = l.WithName("gateway-to-gateway-instance-mapper")

	return func(ctx context.Context, obj kube_client.Object) []kube_reconcile.Request {
		gateway := obj.(*mesh_k8s.MeshGateway)
		l = l.WithValues("gateway", obj.GetName(), "mesh", gateway.Mesh)

		spec, err := gateway.GetSpec()
		if err != nil {
			l.Error(err, "failed to get core resource from MeshGateway")
		}

		var serviceNames []string
		for _, selector := range spec.(*mesh_proto.MeshGateway).GetSelectors() {
			if tagValue, ok := selector.Match[mesh_proto.ServiceTag]; ok {
				serviceNames = append(serviceNames, tagValue)
			}
		}

		var req []kube_reconcile.Request
		for _, serviceName := range serviceNames {
			instances := &mesh_k8s.MeshGatewayInstanceList{}
			if err := client.List(
				ctx, instances, kube_client.MatchingFields{serviceKey: serviceName},
			); err != nil {
				l.Error(err, "failed to fetch GatewayInstances")
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
		Watches(&mesh_k8s.MeshGateway{}, kube_handler.EnqueueRequestsFromMapFunc(GatewayToInstanceMapper(r.Log, mgr.GetClient()))).
		Complete(r)
}
