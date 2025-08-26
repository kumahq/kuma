package gatewayapi

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	kube_core "k8s.io/api/core/v1"
	kube_apierrs "k8s.io/apimachinery/pkg/api/errors"
	kube_apimeta "k8s.io/apimachinery/pkg/api/meta"
	kube_meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	kube_types "k8s.io/apimachinery/pkg/types"
	gatewayapi_v1 "sigs.k8s.io/gateway-api/apis/v1"
	gatewayapi "sigs.k8s.io/gateway-api/apis/v1beta1"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	system_proto "github.com/kumahq/kuma/api/system/v1alpha1"
	mesh_k8s "github.com/kumahq/kuma/pkg/plugins/resources/k8s/native/api/v1alpha1"
	"github.com/kumahq/kuma/pkg/plugins/runtime/k8s/controllers/gatewayapi/common"
	"github.com/kumahq/kuma/pkg/plugins/runtime/k8s/controllers/gatewayapi/referencegrants"
	"github.com/kumahq/kuma/pkg/util/pointer"
)

type ListenerConditions map[gatewayapi.SectionName][]kube_meta.Condition

func validProtocol(crossMesh bool, protocol gatewayapi.ProtocolType) bool {
	switch protocol {
	case gatewayapi_v1.HTTPProtocolType, gatewayapi_v1.HTTPSProtocolType:
		return !crossMesh || protocol == gatewayapi_v1.HTTPProtocolType
	default:
	}

	return false
}

func ValidateListeners(crossMesh bool, listeners []gatewayapi.Listener) ([]gatewayapi.Listener, ListenerConditions) {
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
				Type:    string(gatewayapi_v1.ListenerConditionAccepted),
				Status:  kube_meta.ConditionFalse,
				Reason:  string(reason),
				Message: message,
			},
			kube_meta.Condition{
				Type:    string(gatewayapi_v1.ListenerConditionProgrammed),
				Status:  kube_meta.ConditionFalse,
				Reason:  string(gatewayapi_v1.ListenerReasonInvalid),
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
				Type:    string(gatewayapi_v1.ListenerConditionConflicted),
				Status:  kube_meta.ConditionTrue,
				Reason:  string(reason),
				Message: message,
			},
			kube_meta.Condition{
				Type:    string(gatewayapi_v1.ListenerConditionProgrammed),
				Status:  kube_meta.ConditionFalse,
				Reason:  string(gatewayapi_v1.ListenerReasonInvalid),
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

		if validProtocol(crossMesh, l.Protocol) {
			protocols[l.Protocol] = struct{}{}
		}
		portProtocols[l.Port] = protocols
	}

	for _, l := range listeners {
		if !validProtocol(crossMesh, l.Protocol) {
			message := fmt.Sprintf("unsupported protocol %s", l.Protocol)
			if crossMesh {
				message = fmt.Sprintf("%s with cross-mesh", message)
			}
			appendDetachedCondition(
				l.Name,
				gatewayapi_v1.ListenerReasonUnsupportedProtocol,
				message,
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
					gatewayapi_v1.ListenerReasonHostnameConflict,
					fmt.Sprintf("multiple listeners for %s:%d", *hn, l.Port),
				)
				continue
			}
		}

		if protocols := portProtocols[l.Port]; len(protocols) > 1 {
			appendConflictedCondition(
				l.Name,
				gatewayapi_v1.ListenerReasonProtocolConflict,
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
	mesh string,
	gateway *gatewayapi.Gateway,
	config mesh_k8s.MeshGatewayConfigSpec,
) (*mesh_proto.MeshGateway, ListenerConditions, error) {
	validListeners, listenerConditions := ValidateListeners(config.CrossMesh, gateway.Spec.Listeners)

	var listeners []*mesh_proto.MeshGateway_Listener

	for _, l := range validListeners {
		listener := &mesh_proto.MeshGateway_Listener{
			Port: uint32(l.Port),
			Tags: map[string]string{
				// gateway-api routes are configured using direct references to
				// Gateways, so just create a tag specifically for this listener
				mesh_proto.ListenerTag: string(l.Name),
				mesh_proto.PortTag:     strconv.Itoa(int(l.Port)),
			},
			CrossMesh: config.CrossMesh,
		}

		if protocol, ok := mesh_proto.MeshGateway_Listener_Protocol_value[string(l.Protocol)]; ok {
			listener.Protocol = mesh_proto.MeshGateway_Listener_Protocol(protocol)
		} else if l.Protocol != "" {
			// TODO admission webhook should prevent this
			listenerConditions[l.Name] = append(listenerConditions[l.Name],
				kube_meta.Condition{
					Type:    string(gatewayapi_v1.ListenerConditionProgrammed),
					Status:  kube_meta.ConditionFalse,
					Reason:  string(gatewayapi_v1.ListenerReasonInvalid),
					Message: fmt.Sprintf("unexpected protocol %s", l.Protocol),
				},
			)
			continue
		}

		var unsupportedRouteGroupKinds []string
		for _, gk := range l.AllowedRoutes.Kinds {
			if gk.Kind != common.HTTPRouteKind || *gk.Group != gatewayapi.GroupName {
				metaGK := kube_meta.GroupKind{Group: string(*gk.Group), Kind: string(gk.Kind)}
				unsupportedRouteGroupKinds = append(unsupportedRouteGroupKinds, metaGK.String())
			}
		}
		if len(unsupportedRouteGroupKinds) > 0 {
			listenerConditions[l.Name] = append(listenerConditions[l.Name],
				kube_meta.Condition{
					Type:    string(gatewayapi_v1.ListenerConditionResolvedRefs),
					Status:  kube_meta.ConditionFalse,
					Reason:  string(gatewayapi_v1.ListenerReasonInvalidRouteKinds),
					Message: fmt.Sprintf("unexpected RouteGroupKind %q", strings.Join(unsupportedRouteGroupKinds, ", ")),
				},
			)
		}

		listener.Hostname = string(pointer.DerefOr(l.Hostname, mesh_proto.WildcardHostname))

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
		match := map[string]string{
			mesh_proto.ServiceTag: fmt.Sprintf("%s_%s_svc", gateway.Name, gateway.Namespace),
		}

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

func (r *GatewayReconciler) handleCertRefs(ctx context.Context, mesh, gatewayNamespace string, l gatewayapi.Listener) (*mesh_proto.MeshGateway_TLS_Conf, *certRefCondition, error) {
	type referencedSecret struct {
		name kube_types.NamespacedName
		data []byte
	}
	var referencedSecrets []referencedSecret
	for _, certRef := range l.TLS.CertificateRefs {
		policyRef := referencegrants.PolicyReferenceSecret(referencegrants.FromGatewayIn(gatewayNamespace), certRef)

		name := fmt.Sprintf("%q %q", policyRef.GroupKindReferredTo().String(), policyRef.NamespacedNameReferredTo().String())
		if permitted, err := referencegrants.IsReferencePermitted(ctx, r.Client, policyRef); err != nil {
			return nil, nil, err
		} else if !permitted {
			return nil, &certRefCondition{
				reason:  string(gatewayapi_v1.ListenerReasonRefNotPermitted),
				message: fmt.Sprintf("reference to %s not permitted by any ReferenceGrant", name),
			}, nil
		}

		if *certRef.Kind != "Secret" || *certRef.Group != "" {
			return nil, &certRefCondition{
				reason:  string(gatewayapi_v1.ListenerReasonInvalidCertificateRef),
				message: fmt.Sprintf("invalid reference to %s/%s", *certRef.Group, *certRef.Kind),
			}, nil
		}

		secret := &kube_core.Secret{}
		if err := r.Get(ctx, policyRef.NamespacedNameReferredTo(), secret); err != nil {
			if kube_apierrs.IsNotFound(err) {
				return nil, &certRefCondition{
					reason:  string(gatewayapi_v1.ListenerReasonInvalidCertificateRef),
					message: fmt.Sprintf("invalid reference to %s", name),
				}, nil
			}

			return nil, nil, err
		}

		data, err := convertSecret(secret)
		if err != nil {
			return nil, &certRefCondition{
				reason:  string(gatewayapi_v1.ListenerReasonInvalidCertificateRef),
				message: err.Error(),
			}, nil
		}

		referencedSecrets = append(referencedSecrets, referencedSecret{
			name: policyRef.NamespacedNameReferredTo(),
			data: data,
		})
	}

	if l.TLS.Mode != nil && *l.TLS.Mode == gatewayapi_v1.TLSModePassthrough {
		return nil, nil, nil // todo admission webhook should prevent this
	}

	tls := &mesh_proto.MeshGateway_TLS_Conf{
		Mode: mesh_proto.MeshGateway_TLS_TERMINATE,
	}
	for _, secretKey := range referencedSecrets {
		secretKey, err := r.createSecretIfMissing(ctx, secretKey.name, secretKey.data, mesh)
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
			Type:   string(gatewayapi_v1.ListenerConditionAccepted),
			Status: kube_meta.ConditionTrue,
			Reason: string(gatewayapi_v1.ListenerReasonAccepted),
		},
		kube_meta.Condition{
			Type:   string(gatewayapi_v1.ListenerConditionConflicted),
			Status: kube_meta.ConditionFalse,
			Reason: string(gatewayapi_v1.ListenerReasonNoConflicts),
		},
	)

	var resolvedRefConditions []kube_meta.Condition

	if unresolvableCertRef != nil && !kube_apimeta.IsStatusConditionFalse(conditions, string(gatewayapi_v1.ListenerConditionResolvedRefs)) {
		kube_apimeta.SetStatusCondition(&conditions,
			kube_meta.Condition{
				Type:    string(gatewayapi_v1.ListenerConditionResolvedRefs),
				Status:  kube_meta.ConditionFalse,
				Reason:  unresolvableCertRef.reason,
				Message: unresolvableCertRef.message,
			},
		)
	}

	if !kube_apimeta.IsStatusConditionFalse(conditions, string(gatewayapi_v1.ListenerConditionResolvedRefs)) {
		kube_apimeta.SetStatusCondition(&conditions,
			kube_meta.Condition{
				Type:   string(gatewayapi_v1.ListenerConditionResolvedRefs),
				Status: kube_meta.ConditionTrue,
				Reason: string(gatewayapi_v1.ListenerReasonResolvedRefs),
			},
		)
		kube_apimeta.SetStatusCondition(&conditions,
			kube_meta.Condition{
				Type:   string(gatewayapi_v1.ListenerConditionProgrammed),
				Status: kube_meta.ConditionTrue,
				Reason: string(gatewayapi_v1.ListenerReasonProgrammed),
			},
		)
	} else {
		kube_apimeta.SetStatusCondition(&conditions,
			kube_meta.Condition{
				Type:    string(gatewayapi_v1.ListenerConditionProgrammed),
				Status:  kube_meta.ConditionFalse,
				Reason:  string(gatewayapi_v1.ListenerReasonInvalid),
				Message: "unable to resolve refs",
			},
		)
	}

	return append(conditions, resolvedRefConditions...)
}
