package gatewayapi

import (
	"context"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	kube_apierrs "k8s.io/apimachinery/pkg/api/errors"
	kube_runtime "k8s.io/apimachinery/pkg/runtime"
	kube_types "k8s.io/apimachinery/pkg/types"
	kube_ctrl "sigs.k8s.io/controller-runtime"
	kube_client "sigs.k8s.io/controller-runtime/pkg/client"
	gatewayapi "sigs.k8s.io/gateway-api/apis/v1alpha2"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/manager"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/store"
	k8s_common "github.com/kumahq/kuma/pkg/plugins/common/k8s"
	k8s_util "github.com/kumahq/kuma/pkg/plugins/runtime/k8s/util"
	util_k8s "github.com/kumahq/kuma/pkg/util/k8s"
)

// HTTPRouteReconciler reconciles a GatewayAPI object into Kuma-native objects
type HTTPRouteReconciler struct {
	kube_client.Client
	Log logr.Logger

	Scheme          *kube_runtime.Scheme
	Converter       k8s_common.Converter
	SystemNamespace string
	ResourceManager manager.ResourceManager
}

// Reconcile handles transforming a gateway-api HTTPRoute into a Kuma
// GatewayRoute and managing the status of the gateway-api objects.
func (r *HTTPRouteReconciler) Reconcile(ctx context.Context, req kube_ctrl.Request) (kube_ctrl.Result, error) {
	httpRoute := &gatewayapi.HTTPRoute{}
	if err := r.Get(ctx, req.NamespacedName, httpRoute); err != nil {
		if kube_apierrs.IsNotFound(err) {
			return kube_ctrl.Result{}, nil
		}

		return kube_ctrl.Result{}, err
	}

	// TODO check that gateways exist etc
	// TODO set status on gapi resources

	coreName := util_k8s.K8sNamespacedNameToCoreName(httpRoute.Name, httpRoute.Namespace)
	mesh := k8s_util.MeshFor(httpRoute)

	resource := core_mesh.NewGatewayRouteResource()

	if err := manager.Upsert(r.ResourceManager, model.ResourceKey{Mesh: mesh, Name: coreName}, resource, func(resource model.Resource) error {
		spec, err := r.gapiToKumaRoute(ctx, mesh, httpRoute.Namespace, httpRoute)
		if err != nil {
			return errors.Wrap(err, "error generating GatewayRoute")
		}

		return resource.SetSpec(spec)
	}); err != nil {
		return kube_ctrl.Result{}, errors.Wrap(err, "could not upsert GatewayRoute")
	}

	var err error

	resource.Spec, err = r.gapiToKumaRoute(ctx, mesh, httpRoute.Namespace, httpRoute)
	if err != nil {
		return kube_ctrl.Result{}, err
	}

	err = r.ResourceManager.Update(ctx, resource, store.ModifiedAt(core.Now()))

	return kube_ctrl.Result{}, errors.Wrap(err, "could not update GatewayRoute resource")
}

func (r *HTTPRouteReconciler) gapiToKumaRoute(ctx context.Context, mesh string, namespace string, route *gatewayapi.HTTPRoute) (*mesh_proto.GatewayRoute, error) {
	var selectors []*mesh_proto.Selector

	// Convert GAPI parent refs into Kuma tag matchers
	for _, ref := range route.Spec.ParentRefs {
		namespace := route.Namespace
		if ns := ref.Namespace; ns != nil {
			namespace = string(*ns)
		}

		match := serviceTagForGateway(kube_types.NamespacedName{Namespace: namespace, Name: string(ref.Name)})

		if ref.SectionName != nil {
			match[mesh_proto.ListenerTag] = string(*ref.SectionName)
		}

		selectors = append(selectors, &mesh_proto.Selector{
			Match: match,
		})
	}

	var hostnames []string

	for _, hn := range route.Spec.Hostnames {
		hostnames = append(hostnames, string(hn))
	}

	var rules []*mesh_proto.GatewayRoute_HttpRoute_Rule

	for _, rule := range route.Spec.Rules {
		var backends []*mesh_proto.GatewayRoute_Backend

		for _, backend := range rule.BackendRefs {
			ref := backend.BackendObjectReference

			destination, err := r.gapiToKumaRef(ctx, mesh, namespace, ref)
			if err != nil {
				return nil, err
			}

			backends = append(backends, &mesh_proto.GatewayRoute_Backend{
				// Weight has a default of 1
				Weight:      uint32(*backend.Weight),
				Destination: destination,
			})
		}

		var matches []*mesh_proto.GatewayRoute_HttpRoute_Match

		for _, match := range rule.Matches {
			kumaMatch, err := gapiToKumaMatch(match)
			if err != nil {
				return nil, errors.Wrap(err, "couldn't convert match")
			}

			matches = append(matches, kumaMatch)
		}

		var filters []*mesh_proto.GatewayRoute_HttpRoute_Filter

		for _, filter := range rule.Filters {
			kumaFilter, err := r.gapiToKumaFilter(ctx, mesh, namespace, filter)
			if err != nil {
				return nil, err
			}

			filters = append(filters, kumaFilter)
		}

		rules = append(rules, &mesh_proto.GatewayRoute_HttpRoute_Rule{
			Matches:  matches,
			Filters:  filters,
			Backends: backends,
		})
	}

	return &mesh_proto.GatewayRoute{
		Selectors: selectors,
		Conf: &mesh_proto.GatewayRoute_Conf{
			Route: &mesh_proto.GatewayRoute_Conf_Http{
				Http: &mesh_proto.GatewayRoute_HttpRoute{
					Hostnames: hostnames,
					Rules:     rules,
				},
			},
		},
	}, nil
}

func (r *HTTPRouteReconciler) SetupWithManager(mgr kube_ctrl.Manager) error {
	return kube_ctrl.NewControllerManagedBy(mgr).
		For(&gatewayapi.HTTPRoute{}).
		Complete(r)
}
