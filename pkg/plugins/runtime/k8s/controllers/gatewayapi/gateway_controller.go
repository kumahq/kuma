package gatewayapi

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	kube_core "k8s.io/api/core/v1"
	kube_apierrs "k8s.io/apimachinery/pkg/api/errors"
	kube_meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	kube_runtime "k8s.io/apimachinery/pkg/runtime"
	kube_schema "k8s.io/apimachinery/pkg/runtime/schema"
	kube_ctrl "sigs.k8s.io/controller-runtime"
	kube_client "sigs.k8s.io/controller-runtime/pkg/client"
	kube_controllerutil "sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	gatewayapi "sigs.k8s.io/gateway-api/apis/v1alpha2"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/resources/manager"
	mesh_k8s "github.com/kumahq/kuma/pkg/plugins/resources/k8s/native/api/v1alpha1"
	k8s_registry "github.com/kumahq/kuma/pkg/plugins/resources/k8s/native/pkg/registry"
	"github.com/kumahq/kuma/pkg/plugins/runtime/k8s/containers"
	k8s_util "github.com/kumahq/kuma/pkg/plugins/runtime/k8s/util"
	util_k8s "github.com/kumahq/kuma/pkg/util/k8s"
	"github.com/kumahq/kuma/pkg/plugins/runtime/k8s/controllers/gatewayapi/policy"
)

// GatewayReconciler reconciles a GatewayAPI Gateway object.
type GatewayReconciler struct {
	kube_client.Client
	Log logr.Logger

	Scheme          *kube_runtime.Scheme
	TypeRegistry    k8s_registry.TypeRegistry
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
			err := reconcileLabelledObject(ctx, r.TypeRegistry, r.Client, req.NamespacedName, &mesh_proto.Gateway{}, nil)
			return kube_ctrl.Result{}, errors.Wrap(err, "could not delete owned Gateway.kuma.io")
		}

		return kube_ctrl.Result{}, err
	}

	class, err := getGatewayClass(ctx, r.Client, gateway.Spec.GatewayClassName)
	if err != nil {
		return kube_ctrl.Result{}, errors.Wrap(err, "unable to retrieve GatewayClass referenced by Gateway")
	}

	if class.Spec.ControllerName != controllerName {
		return kube_ctrl.Result{}, nil
	}

	gatewaySpec, listenerConditions, err := r.gapiToKumaGateway(ctx, gateway)
	if err != nil {
		return kube_ctrl.Result{}, errors.Wrap(err, "error generating Gateway.kuma.io")
	}

	if err := reconcileLabelledObject(ctx, r.TypeRegistry, r.Client, req.NamespacedName, &mesh_proto.Gateway{}, gatewaySpec); err != nil {
		return kube_ctrl.Result{}, errors.Wrap(err, "could not reconcile owned Gateway.kuma.io")
	}

	gatewayInstance, err := r.createOrUpdateInstance(ctx, gateway)
	if err != nil {
		return kube_ctrl.Result{}, errors.Wrap(err, "unable to reconcile GatewayInstance")
	}

	if err := r.updateStatus(ctx, gateway, gatewayInstance, listenerConditions); err != nil {
		return kube_ctrl.Result{}, errors.Wrap(err, "unable to update Gateway status")
	}

	return kube_ctrl.Result{}, nil
}

type ListenerConditions map[gatewayapi.SectionName][]kube_meta.Condition

func (r *GatewayReconciler) gapiToKumaGateway(
	ctx context.Context, gateway *gatewayapi.Gateway,
) (*mesh_proto.Gateway, ListenerConditions, error) {
	var listeners []*mesh_proto.Gateway_Listener

	listenerConditions := ListenerConditions{}

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
			// TODO admission webhook should prevent this
			listenerConditions[l.Name] = append(listenerConditions[l.Name],
				kube_meta.Condition{
					Type:    string(gatewayapi.ListenerConditionReady),
					Status:  kube_meta.ConditionFalse,
					Reason:  string(gatewayapi.ListenerReasonInvalid),
					Message: fmt.Sprintf("unexpected protocol %s", l.Protocol),
				},
			)
			continue
		}

		listener.Hostname = "*"
		if l.Hostname != nil {
			listener.Hostname = string(*l.Hostname)
		}

		refsPermitted := true

		for _, certRef := range l.TLS.CertificateRefs {
			ref := policy.PolicyReferenceSecret(policy.FromGatewayIn(gateway.Namespace), *certRef)

			permitted, err := policy.IsReferencePermitted(ctx, r.Client, ref)
			if err != nil {
				return nil, nil, err
			}

			refsPermitted = refsPermitted && permitted

			var resolvedCondition []kube_meta.Condition

			if permitted {
				resolvedCondition = []kube_meta.Condition{
					{
						Type:   string(gatewayapi.ListenerConditionResolvedRefs),
						Status: kube_meta.ConditionTrue,
						Reason: string(gatewayapi.ListenerReasonResolvedRefs),
					},
					{
						Type:   string(gatewayapi.ListenerConditionReady),
						Status: kube_meta.ConditionTrue,
						Reason: string(gatewayapi.ListenerConditionReady),
					},
				}
			} else {
				targetGK := kube_schema.GroupKind{Group: string(*certRef.Group), Kind: string(*certRef.Kind)}
				resolvedCondition = []kube_meta.Condition{
					{
						Type:    string(gatewayapi.ListenerConditionResolvedRefs),
						Status:  kube_meta.ConditionFalse,
						Reason:  string(gatewayapi.ListenerReasonRefNotPermitted),
						Message: fmt.Sprintf("reference to %s %s/%s not permitted by any ReferencePolicy", targetGK.String(), ref.ToNamespace(), certRef.Name),
					},
					{
						Type:    string(gatewayapi.ListenerConditionReady),
						Status:  kube_meta.ConditionFalse,
						Reason:  string(gatewayapi.ListenerReasonInvalid),
						Message: "unable to resolve refs",
					},
				}
			}

			listenerConditions[l.Name] = append(listenerConditions[l.Name], resolvedCondition...)
		}

		if refsPermitted {
			listeners = append(listeners, listener)
		}
	}

	match := serviceTagForGateway(kube_client.ObjectKeyFromObject(gateway))

	return &mesh_proto.Gateway{
		Selectors: []*mesh_proto.Selector{
			{Match: match},
		},
		Conf: &mesh_proto.Gateway_Conf{
			Listeners: listeners,
		},
	}, listenerConditions, nil
}

func (r *GatewayReconciler) createOrUpdateInstance(ctx context.Context, gateway *gatewayapi.Gateway) (*mesh_k8s.GatewayInstance, error) {
	instance := &mesh_k8s.GatewayInstance{
		ObjectMeta: kube_meta.ObjectMeta{
			Namespace: gateway.Namespace,
			Name:      gateway.Name,
		},
	}

	if _, err := kube_controllerutil.CreateOrUpdate(ctx, r.Client, instance, func() error {
		instance.Spec = mesh_k8s.GatewayInstanceSpec{
			ServiceType: kube_core.ServiceTypeLoadBalancer,
			Tags:        serviceTagForGateway(kube_client.ObjectKeyFromObject(gateway)),
		}

		err := kube_controllerutil.SetControllerReference(gateway, instance, r.Scheme)
		return errors.Wrap(err, "unable to set GatewayInstance's controller reference to Gateway")
	}); err != nil {
		return nil, errors.Wrap(err, "couldn't create GatewayInstance")
	}

	return instance, nil
}

func (r *GatewayReconciler) SetupWithManager(mgr kube_ctrl.Manager) error {
	return kube_ctrl.NewControllerManagedBy(mgr).
		For(&gatewayapi.Gateway{}).
		Owns(&mesh_k8s.Gateway{}).
		Owns(&mesh_k8s.GatewayInstance{}).
		Complete(r)
}
