package gatewayapi

import (
	"context"
	"fmt"
	"strings"

	kube_meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	kube_client "sigs.k8s.io/controller-runtime/pkg/client"
	gatewayapi "sigs.k8s.io/gateway-api/apis/v1alpha2"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	system_proto "github.com/kumahq/kuma/api/system/v1alpha1"
	"github.com/kumahq/kuma/pkg/plugins/runtime/k8s/controllers/gatewayapi/common"
	"github.com/kumahq/kuma/pkg/plugins/runtime/k8s/controllers/gatewayapi/policy"
)

type ListenerConditions map[gatewayapi.SectionName][]kube_meta.Condition

func validProtocol(protocol gatewayapi.ProtocolType) bool {
	switch protocol {
	case gatewayapi.HTTPProtocolType, gatewayapi.HTTPSProtocolType:
		return true
	default:
	}

	return false
}

func ValidateListeners(listeners []gatewayapi.Listener) ([]gatewayapi.Listener, ListenerConditions) {
	var validListeners []gatewayapi.Listener
	listenerConditions := ListenerConditions{}

	appendDetachedCondition := func(
		listener gatewayapi.SectionName,
		reason gatewayapi.ListenerConditionReason,
		message string,
	) {
		listenerConditions[listener] = append(
			listenerConditions[listener],
			kube_meta.Condition{
				Type:    string(gatewayapi.ListenerConditionDetached),
				Status:  kube_meta.ConditionTrue,
				Reason:  string(reason),
				Message: message,
			},
			kube_meta.Condition{
				Type:    string(gatewayapi.ListenerConditionReady),
				Status:  kube_meta.ConditionFalse,
				Reason:  string(gatewayapi.ListenerReasonInvalid),
				Message: "detached",
			},
		)
	}

	appendConflictedCondition := func(
		listener gatewayapi.SectionName,
		reason gatewayapi.ListenerConditionReason,
		message string,
	) {
		listenerConditions[listener] = append(
			listenerConditions[listener],
			kube_meta.Condition{
				Type:    string(gatewayapi.ListenerConditionConflicted),
				Status:  kube_meta.ConditionTrue,
				Reason:  string(reason),
				Message: message,
			},
			kube_meta.Condition{
				Type:    string(gatewayapi.ListenerConditionReady),
				Status:  kube_meta.ConditionFalse,
				Reason:  string(gatewayapi.ListenerReasonInvalid),
				Message: "conflicts found",
			},
		)
	}

	// Collect information about used hostnames and protocols
	// We can only have at most one listener for each HostnamePort
	type HostnamePort struct {
		hostname gatewayapi.Hostname
		port     gatewayapi.PortNumber
	}
	portHostnames := map[HostnamePort]int{}
	// We can only have one protocol for each port
	portProtocols := map[gatewayapi.PortNumber]map[gatewayapi.ProtocolType]struct{}{}

	for _, l := range listeners {
		if hn := l.Hostname; hn != nil {
			hostnamePort := HostnamePort{
				hostname: *hn,
				port:     l.Port,
			}
			portHostnames[hostnamePort]++
		}

		protocols, ok := portProtocols[l.Port]
		if !ok {
			protocols = map[gatewayapi.ProtocolType]struct{}{}
		}

		if validProtocol(l.Protocol) {
			protocols[l.Protocol] = struct{}{}
		}
		portProtocols[l.Port] = protocols
	}

	for _, l := range listeners {
		if !validProtocol(l.Protocol) {
			appendDetachedCondition(
				l.Name,
				gatewayapi.ListenerReasonUnsupportedProtocol,
				fmt.Sprintf("unsupported protocol %s", l.Protocol),
			)
			continue
		}

		// TODO ListenerReasonUnsupportedAddress and ListenerReasonPortUnavailable
		// need more information from Envoy Gateway

		if hn := l.Hostname; hn != nil {
			hostnamePort := HostnamePort{
				hostname: *hn,
				port:     l.Port,
			}
			if num := portHostnames[hostnamePort]; num > 1 {
				appendConflictedCondition(
					l.Name,
					gatewayapi.ListenerReasonHostnameConflict,
					fmt.Sprintf("multiple listeners for %s:%d", *hn, l.Port),
				)
				continue
			}
		}

		if protocols := portProtocols[l.Port]; len(protocols) > 1 {
			appendConflictedCondition(
				l.Name,
				gatewayapi.ListenerReasonProtocolConflict,
				fmt.Sprintf("multiple listeners on %d with conflicting protocols", l.Port),
			)
			continue
		}

		// We don't set ListenerReasonRouteConflict because we already check the
		// routes with ListenerReasonInvalidRouteKinds
		// Once we support more than HTTPRoute it may be fitting to set this
		// depending on the listener protocol

		validListeners = append(validListeners, l)
	}

	return validListeners, listenerConditions
}

// gapiToKumaGateway returns a converted gateway (if possible) and any
// conditions to set on the gatewayapi listeners
func (r *GatewayReconciler) gapiToKumaGateway(
	ctx context.Context,
	gateway *gatewayapi.Gateway,
	mesh string,
) (*mesh_proto.MeshGateway, ListenerConditions, error) {
	validListeners, listenerConditions := ValidateListeners(gateway.Spec.Listeners)

	var listeners []*mesh_proto.MeshGateway_Listener

	for _, l := range validListeners {
		listener := &mesh_proto.MeshGateway_Listener{
			Port: uint32(l.Port),
			Tags: map[string]string{
				// gateway-api routes are configured using direct references to
				// Gateways, so just create a tag specifically for this listener
				mesh_proto.ListenerTag: string(l.Name),
			},
		}

		if protocol, ok := mesh_proto.MeshGateway_Listener_Protocol_value[string(l.Protocol)]; ok {
			listener.Protocol = mesh_proto.MeshGateway_Listener_Protocol(protocol)
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

		for _, gk := range l.AllowedRoutes.Kinds {
			if gk.Kind != common.HTTPRouteKind || *gk.Group != gatewayapi.GroupName {
				metaGK := kube_meta.GroupKind{Group: string(*gk.Group), Kind: string(gk.Kind)}
				listenerConditions[l.Name] = append(listenerConditions[l.Name],
					kube_meta.Condition{
						Type:    string(gatewayapi.ListenerConditionResolvedRefs),
						Status:  kube_meta.ConditionFalse,
						Reason:  string(gatewayapi.ListenerReasonInvalidRouteKinds),
						Message: fmt.Sprintf("unexpected RouteGroupKind %q", metaGK.String()),
					},
				)
				continue
			}
		}

		listener.Hostname = "*"
		if l.Hostname != nil {
			listener.Hostname = string(*l.Hostname)
		}

		var unresolvableRefs []string
		if l.TLS != nil {
			for _, certRef := range l.TLS.CertificateRefs {
				policyRef := policy.PolicyReferenceSecret(policy.FromGatewayIn(gateway.Namespace), certRef)

				permitted, err := policy.IsReferencePermitted(ctx, r.Client, policyRef)
				if err != nil {
					return nil, nil, err
				}

				if !permitted {
					message := fmt.Sprintf("%q %q", policyRef.GroupKindReferredTo().String(), policyRef.NamespacedNameReferredTo().String())
					unresolvableRefs = append(unresolvableRefs, message)
				}
			}

			if len(unresolvableRefs) == 0 {
				if l.TLS.Mode != nil && *l.TLS.Mode == gatewayapi.TLSModePassthrough {
					continue // todo admission webhook should prevent this
				}
				listener.Tls = &mesh_proto.MeshGateway_TLS_Conf{
					Mode: mesh_proto.MeshGateway_TLS_TERMINATE,
				}
				for _, certRef := range l.TLS.CertificateRefs {
					// We only support canonical Secret of TLS type. Ignore other types
					namespacedName := types.NamespacedName{
						Name:      string(certRef.Name),
						Namespace: gateway.Namespace, // if not specified, namespace is inherited from Gateway
					}
					if certRef.Namespace != nil {
						namespacedName.Namespace = string(*certRef.Namespace)
					}
					secretKey, err := r.createSecretIfMissing(ctx, namespacedName, mesh)
					if err != nil {
						return nil, nil, err
					}
					listener.Tls.Certificates = append(listener.Tls.Certificates, &system_proto.DataSource{
						Type: &system_proto.DataSource_Secret{
							Secret: secretKey.Name,
						},
					})
				}
			}
		}

		// We've already cleared this listener of conflicts and being detached
		listenerConditions[l.Name] = append(
			listenerConditions[l.Name],
			kube_meta.Condition{
				Type:   string(gatewayapi.ListenerConditionDetached),
				Status: kube_meta.ConditionFalse,
				Reason: string(gatewayapi.ListenerReasonAttached),
			},
			kube_meta.Condition{
				Type:   string(gatewayapi.ListenerConditionConflicted),
				Status: kube_meta.ConditionFalse,
				Reason: string(gatewayapi.ListenerReasonNoConflicts),
			},
		)

		var resolvedRefConditions []kube_meta.Condition

		if len(unresolvableRefs) == 0 {
			listeners = append(listeners, listener)

			resolvedRefConditions = []kube_meta.Condition{
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
			resolvedRefConditions = []kube_meta.Condition{
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

		listenerConditions[l.Name] = append(listenerConditions[l.Name], resolvedRefConditions...)
	}

	var kumaGateway *mesh_proto.MeshGateway

	if len(listeners) > 0 {
		match := common.ServiceTagForGateway(kube_client.ObjectKeyFromObject(gateway))

		kumaGateway = &mesh_proto.MeshGateway{
			Selectors: []*mesh_proto.Selector{
				{Match: match},
			},
			Conf: &mesh_proto.MeshGateway_Conf{
				Listeners: listeners,
			},
		}
	}

	return kumaGateway, listenerConditions, nil
}
