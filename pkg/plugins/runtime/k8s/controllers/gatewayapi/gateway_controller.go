package gatewayapi

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	kube_apps "k8s.io/api/apps/v1"
	kube_core "k8s.io/api/core/v1"
	kube_apierrs "k8s.io/apimachinery/pkg/api/errors"
	kube_runtime "k8s.io/apimachinery/pkg/runtime"
	kube_types "k8s.io/apimachinery/pkg/types"
	kube_ctrl "sigs.k8s.io/controller-runtime"
	kube_client "sigs.k8s.io/controller-runtime/pkg/client"
	gatewayapi "sigs.k8s.io/gateway-api/apis/v1alpha2"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/manager"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	k8s_common "github.com/kumahq/kuma/pkg/plugins/common/k8s"
	mesh_k8s "github.com/kumahq/kuma/pkg/plugins/resources/k8s/native/api/v1alpha1"
	"github.com/kumahq/kuma/pkg/plugins/runtime/k8s/containers"
	k8s_util "github.com/kumahq/kuma/pkg/plugins/runtime/k8s/util"
	util_k8s "github.com/kumahq/kuma/pkg/util/k8s"
)

const controllerName = "gateways.kuma.io/controller"

// GatewayReconciler reconciles a GatewayAPI Gateway object.
type GatewayReconciler struct {
	kube_client.Client
	Reader kube_client.Reader
	Log    logr.Logger

	Scheme          *kube_runtime.Scheme
	Converter       k8s_common.Converter
	SystemNamespace string
	ProxyFactory    containers.DataplaneProxyFactory
	ResourceManager manager.ResourceManager
}

// Reconcile handles transforming a gateway-api Gateway into a Kuma Gateway and
// managing the status of the gateway-api objects.
func (r *GatewayReconciler) Reconcile(ctx context.Context, req kube_ctrl.Request) (kube_ctrl.Result, error) {
	gateway := &gatewayapi.Gateway{}
	if err := r.Get(ctx, req.NamespacedName, gateway); err != nil {
		if kube_apierrs.IsNotFound(err) {
			return kube_ctrl.Result{}, nil
		}

		return kube_ctrl.Result{}, err
	}

	class, err := r.getGatewayClass(ctx, gateway.Spec.GatewayClassName)
	if err != nil {
		return kube_ctrl.Result{}, errors.Wrap(err, "unable to retrieve GatewayClass referenced by Gateway")
	} else if class == nil {
		return kube_ctrl.Result{}, nil
	}

	coreName := util_k8s.K8sNamespacedNameToCoreName(gateway.Name, gateway.Namespace)
	mesh := k8s_util.MeshFor(gateway)

	resource := core_mesh.NewGatewayResource()

	if err := manager.Upsert(r.ResourceManager, model.ResourceKey{Mesh: mesh, Name: coreName}, resource, func(resource model.Resource) error {
		gatewaySpec, err := r.gapiToKumaGateway(gateway)
		if err != nil {
			return errors.Wrap(err, "could not create Kuma Gateway spec")
		}

		return resource.SetSpec(gatewaySpec)
	}); err != nil {
		return kube_ctrl.Result{}, errors.Wrap(err, "could not upsert Gateway")
	}

	deployment, err := r.createOrUpdateDeployment(ctx, resource, gateway)
	if err != nil {
		return kube_ctrl.Result{}, errors.Wrap(err, "unable to create Deployment for Gateway")
	}

	svc, err := r.createOrUpdateService(ctx, resource, gateway)
	if err != nil {
		return kube_ctrl.Result{}, errors.Wrap(err, "unable to create Service for Gateway")
	}

	r.updateStatus(gateway, svc, deployment)

	if err := r.Get(ctx, req.NamespacedName, gateway); err != nil {
		return kube_ctrl.Result{}, errors.Wrap(err, "unable to get Gateway")
	}

	if err := r.Client.Status().Update(ctx, gateway); err != nil {
		return kube_ctrl.Result{}, errors.Wrap(err, "unable to update Gateway status")
	}

	return kube_ctrl.Result{}, nil
}

func (r *GatewayReconciler) getGatewayClass(ctx context.Context, name gatewayapi.ObjectName) (*gatewayapi.GatewayClass, error) {
	class := &gatewayapi.GatewayClass{}
	classObjectKey := kube_types.NamespacedName{Name: string(name)}

	if err := r.Client.Get(ctx, classObjectKey, class); err != nil {
		if kube_apierrs.IsNotFound(err) {
			return nil, nil
		}

		return nil, errors.Wrapf(err, "failed to get GatewayClass %s", classObjectKey)
	}

	if class.Spec.ControllerName != controllerName {
		return nil, nil
	}

	return class, nil
}

func (r *GatewayReconciler) gapiToKumaGateway(gateway *gatewayapi.Gateway) (*mesh_proto.Gateway, error) {
	var listeners []*mesh_proto.Gateway_Listener

	for _, l := range gateway.Spec.Listeners {
		listener := &mesh_proto.Gateway_Listener{
			Port: uint32(l.Port),
			Tags: map[string]string{
				// gateway-api routes are configured using direct references to
				// Gateways, so just create a tag specifically for this listener
				mesh_proto.ListenerTag: string(l.Name),
			},
		}

		if protocol, ok := mesh_proto.Gateway_Listener_Protocol_value[string(l.Protocol)]; ok {
			listener.Protocol = mesh_proto.Gateway_Listener_Protocol(protocol)
		} else if l.Protocol != "" {
			return nil, errors.Errorf("unexpected protocol %s", l.Protocol)
		}

		listener.Hostname = "*"
		if l.Hostname != nil {
			listener.Hostname = string(*l.Hostname)
		}

		listeners = append(listeners, listener)
	}

	match := map[string]string{
		mesh_proto.ServiceTag: fmt.Sprintf("%s-kuma-gateway_%s_svc", gateway.Name, gateway.Namespace),
	}

	return &mesh_proto.Gateway{
		Selectors: []*mesh_proto.Selector{
			{Match: match},
		},
		Conf: &mesh_proto.Gateway_Conf{
			Listeners: listeners,
		},
	}, nil
}

func (r *GatewayReconciler) updateStatus(gateway *gatewayapi.Gateway, svc *kube_core.Service, deployment *kube_apps.Deployment) {
	ipType := gatewayapi.IPAddressType

	var addrs []gatewayapi.GatewayAddress

	for _, addr := range svc.Status.LoadBalancer.Ingress {
		addrs = append(addrs, gatewayapi.GatewayAddress{
			Type:  &ipType,
			Value: addr.IP,
		})
	}

	gateway.Status.Addresses = addrs

	setConditions(gateway, deployment)
}

func (r *GatewayReconciler) SetupWithManager(mgr kube_ctrl.Manager) error {
	return kube_ctrl.NewControllerManagedBy(mgr).
		For(&gatewayapi.Gateway{}).
		Owns(&mesh_k8s.Gateway{}).
		Owns(&kube_core.Service{}).
		Owns(&kube_apps.Deployment{}).
		Complete(r)
}
