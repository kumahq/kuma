package gatewayapi

import (
	"context"
	"fmt"
	"strings"

	kube_meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	kube_client "sigs.k8s.io/controller-runtime/pkg/client"
	gatewayapi "sigs.k8s.io/gateway-api/apis/v1alpha2"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/plugins/runtime/k8s/controllers/gatewayapi/policy"
)

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

		var certificateRefs []*gatewayapi.SecretObjectReference
		if l.TLS != nil {
			certificateRefs = l.TLS.CertificateRefs
		}

		var unresolvableRefs []string

		for _, certRef := range certificateRefs {
			policyRef := policy.PolicyReferenceSecret(policy.FromGatewayIn(gateway.Namespace), *certRef)

			permitted, err := policy.IsReferencePermitted(ctx, r.Client, policyRef)
			if err != nil {
				return nil, nil, err
			}

			if !permitted {
				message := fmt.Sprintf("%q %q", policyRef.GroupKindReferredTo().String(), policyRef.NamespacedNameReferredTo().String())
				unresolvableRefs = append(unresolvableRefs, message)
			}
		}

		var changedListenerConditions []kube_meta.Condition

		if len(unresolvableRefs) == 0 {
			listeners = append(listeners, listener)

			changedListenerConditions = []kube_meta.Condition{
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
			changedListenerConditions = []kube_meta.Condition{
				{
					Type:    string(gatewayapi.ListenerConditionResolvedRefs),
					Status:  kube_meta.ConditionFalse,
					Reason:  string(gatewayapi.ListenerReasonRefNotPermitted),
					Message: fmt.Sprintf("references to %s not permitted by any ReferencePolicy", strings.Join(unresolvableRefs, ", ")),
				},
				{
					Type:    string(gatewayapi.ListenerConditionReady),
					Status:  kube_meta.ConditionFalse,
					Reason:  string(gatewayapi.ListenerReasonInvalid),
					Message: "unable to resolve refs",
				},
			}
		}

		listenerConditions[l.Name] = append(listenerConditions[l.Name], changedListenerConditions...)
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
