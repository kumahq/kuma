package gatewayapi

import (
	"context"
	"fmt"

	kube_core "k8s.io/api/core/v1"
	kube_apierrs "k8s.io/apimachinery/pkg/api/errors"
	kube_meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	kube_types "k8s.io/apimachinery/pkg/types"
	kube_client "sigs.k8s.io/controller-runtime/pkg/client"
	gatewayapi "sigs.k8s.io/gateway-api/apis/v1beta1"

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

		var unresolvableCertRef *certRefCondition
		if l.TLS != nil {
			var err error
			listener.Tls, unresolvableCertRef, err = r.handleCertRefs(ctx, mesh, gateway.Namespace, l)
			if err != nil {
				return nil, nil, err
			}
		}

		if unresolvableCertRef == nil {
			listeners = append(listeners, listener)
		}

		listenerConditions[l.Name] = handleConditions(listenerConditions[l.Name], unresolvableCertRef)
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

type certRefCondition struct {
	message string
	reason  string
}

func (r *GatewayReconciler) handleCertRefs(ctx context.Context, mesh string, gatewayNamespace string, l gatewayapi.Listener) (*mesh_proto.MeshGateway_TLS_Conf, *certRefCondition, error) {
	var referencedSecrets []kube_types.NamespacedName
	for _, certRef := range l.TLS.CertificateRefs {
		policyRef := policy.PolicyReferenceSecret(policy.FromGatewayIn(gatewayNamespace), certRef)

		if permitted, err := policy.IsReferencePermitted(ctx, r.Client, policyRef); err != nil {
			return nil, nil, err
		} else if !permitted {
			name := fmt.Sprintf("%q %q", policyRef.GroupKindReferredTo().String(), policyRef.NamespacedNameReferredTo().String())
			return nil, &certRefCondition{
				reason:  string(gatewayapi.ListenerReasonRefNotPermitted),
				message: fmt.Sprintf("reference to %s not permitted by any ReferenceGrant", name),
			}, nil
		}

		if err := r.Client.Get(ctx, policyRef.NamespacedNameReferredTo(), &kube_core.Secret{}); err != nil {
			if kube_apierrs.IsNotFound(err) {
				name := fmt.Sprintf("%q %q", policyRef.GroupKindReferredTo().String(), policyRef.NamespacedNameReferredTo().String())
				return nil, &certRefCondition{
					reason:  string(gatewayapi.ListenerReasonInvalidCertificateRef),
					message: fmt.Sprintf("invalid reference to %s", name),
				}, nil
			}

			return nil, nil, err
		}

		referencedSecrets = append(referencedSecrets, policyRef.NamespacedNameReferredTo())
	}

	if l.TLS.Mode != nil && *l.TLS.Mode == gatewayapi.TLSModePassthrough {
		return nil, nil, nil // todo admission webhook should prevent this
	}

	tls := &mesh_proto.MeshGateway_TLS_Conf{
		Mode: mesh_proto.MeshGateway_TLS_TERMINATE,
	}
	for _, secretKey := range referencedSecrets {
		// We only support canonical Secret of TLS type. Ignore other types
		secretKey, err := r.createSecretIfMissing(ctx, secretKey, mesh)
		if err != nil {
			return nil, nil, err
		}

		tls.Certificates = append(tls.Certificates, &system_proto.DataSource{
			Type: &system_proto.DataSource_Secret{
				Secret: secretKey.Name,
			},
		})
	}

	return tls, nil, nil
}

func handleConditions(conditions []kube_meta.Condition, unresolvableCertRef *certRefCondition) []kube_meta.Condition {
	// We've already cleared this listener of conflicts and being detached
	conditions = append(
		conditions,
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

	if unresolvableCertRef == nil {
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
				Reason:  unresolvableCertRef.reason,
				Message: unresolvableCertRef.message,
			},
			{
				Type:    string(gatewayapi.ListenerConditionReady),
				Status:  kube_meta.ConditionFalse,
				Reason:  string(gatewayapi.ListenerReasonInvalid),
				Message: "unable to resolve refs",
			},
		}
	}

	return append(conditions, resolvedRefConditions...)
}
