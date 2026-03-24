package gatewayapi

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	kube_meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	gatewayapi_v1 "sigs.k8s.io/gateway-api/apis/v1"
	gatewayapi "sigs.k8s.io/gateway-api/apis/v1beta1"

	"github.com/kumahq/kuma/v2/pkg/plugins/runtime/k8s/controllers/gatewayapi/common"
	"github.com/kumahq/kuma/v2/pkg/util/pointer"
)

var _ = Describe("mergeGatewayListenerStatuses", func() {
	makeGateway := func(listenerNames ...string) *gatewayapi.Gateway {
		same := gatewayapi_v1.NamespacesFromSame
		var listeners []gatewayapi.Listener
		for _, name := range listenerNames {
			listeners = append(listeners, gatewayapi.Listener{
				Name:     gatewayapi.SectionName(name),
				Protocol: gatewayapi_v1.HTTPProtocolType,
				Port:     gatewayapi.PortNumber(80),
				AllowedRoutes: &gatewayapi.AllowedRoutes{
					Namespaces: &gatewayapi.RouteNamespaces{
						From: &same,
					},
				},
			})
		}
		return &gatewayapi.Gateway{
			Spec: gatewayapi.GatewaySpec{
				Listeners: listeners,
			},
		}
	}

	makeConditions := func(names ...string) ListenerConditions {
		conditions := ListenerConditions{}
		for _, name := range names {
			conditions[gatewayapi.SectionName(name)] = []kube_meta.Condition{
				{
					Type:   string(gatewayapi_v1.ListenerConditionAccepted),
					Status: kube_meta.ConditionTrue,
					Reason: string(gatewayapi_v1.ListenerReasonAccepted),
				},
				{
					Type:   string(gatewayapi_v1.ListenerConditionProgrammed),
					Status: kube_meta.ConditionTrue,
					Reason: string(gatewayapi_v1.ListenerReasonProgrammed),
				},
			}
		}
		return conditions
	}

	It("returns listeners sorted by name", func() {
		listenerNames := []string{"zebra", "alpha", "middle", "beta"}
		gateway := makeGateway(listenerNames...)
		conditions := makeConditions(listenerNames...)
		attachedRoutes := AttachedRoutesForListeners{}

		statuses := mergeGatewayListenerStatuses(gateway, conditions, attachedRoutes)

		Expect(statuses).To(HaveLen(4))
		Expect(statuses[0].Name).To(Equal(gatewayapi.SectionName("alpha")))
		Expect(statuses[1].Name).To(Equal(gatewayapi.SectionName("beta")))
		Expect(statuses[2].Name).To(Equal(gatewayapi.SectionName("middle")))
		Expect(statuses[3].Name).To(Equal(gatewayapi.SectionName("zebra")))
	})

	It("preserves attached route counts", func() {
		gateway := makeGateway("b-listener", "a-listener")
		conditions := makeConditions("b-listener", "a-listener")
		attachedRoutes := AttachedRoutesForListeners{
			gatewayapi.SectionName("a-listener"): {num: 3},
			gatewayapi.SectionName("b-listener"): {num: 1},
		}

		statuses := mergeGatewayListenerStatuses(gateway, conditions, attachedRoutes)

		Expect(statuses).To(HaveLen(2))
		// a-listener sorts first
		Expect(statuses[0].Name).To(Equal(gatewayapi.SectionName("a-listener")))
		Expect(statuses[0].AttachedRoutes).To(Equal(int32(3)))
		Expect(statuses[1].Name).To(Equal(gatewayapi.SectionName("b-listener")))
		Expect(statuses[1].AttachedRoutes).To(Equal(int32(1)))
	})
})

var _ = Describe("mergeHTTPRouteStatus", func() {
	makeParentRef := func(namespace, name, sectionName string) gatewayapi.ParentReference {
		group := gatewayapi.Group(gatewayapi.GroupVersion.Group)
		kind := gatewayapi.Kind("Gateway")
		ns := gatewayapi.Namespace(namespace)
		section := gatewayapi.SectionName(sectionName)
		return gatewayapi.ParentReference{
			Group:       &group,
			Kind:        &kind,
			Namespace:   &ns,
			Name:        gatewayapi.ObjectName(name),
			SectionName: &section,
		}
	}

	It("returns parent statuses sorted by parent ref", func() {
		route := &gatewayapi.HTTPRoute{}
		conditions := ParentConditions{
			makeParentRef("ns-b", "gw-z", "listener-1"): {
				{Type: string(gatewayapi.RouteConditionAccepted), Status: kube_meta.ConditionTrue, Reason: "Accepted"},
			},
			makeParentRef("ns-a", "gw-a", "listener-1"): {
				{Type: string(gatewayapi.RouteConditionAccepted), Status: kube_meta.ConditionTrue, Reason: "Accepted"},
			},
			makeParentRef("ns-a", "gw-a", "listener-2"): {
				{Type: string(gatewayapi.RouteConditionAccepted), Status: kube_meta.ConditionTrue, Reason: "Accepted"},
			},
		}

		mergeHTTPRouteStatus(route, conditions)

		Expect(route.Status.Parents).To(HaveLen(3))
		Expect(route.Status.Parents[0].ParentRef.Name).To(Equal(gatewayapi.ObjectName("gw-a")))
		Expect(pointer.Deref(route.Status.Parents[0].ParentRef.SectionName)).To(Equal(gatewayapi.SectionName("listener-1")))
		Expect(route.Status.Parents[1].ParentRef.Name).To(Equal(gatewayapi.ObjectName("gw-a")))
		Expect(pointer.Deref(route.Status.Parents[1].ParentRef.SectionName)).To(Equal(gatewayapi.SectionName("listener-2")))
		Expect(route.Status.Parents[2].ParentRef.Name).To(Equal(gatewayapi.ObjectName("gw-z")))
	})

	It("preserves statuses from other controllers unsorted", func() {
		otherController := gatewayapi.GatewayController("other-controller")
		route := &gatewayapi.HTTPRoute{
			Status: gatewayapi.HTTPRouteStatus{
				RouteStatus: gatewayapi.RouteStatus{
					Parents: []gatewayapi.RouteParentStatus{
						{
							ParentRef:      makeParentRef("ns-z", "gw-z", "listener-1"),
							ControllerName: otherController,
						},
						{
							ParentRef:      makeParentRef("ns-a", "gw-a", "listener-1"),
							ControllerName: otherController,
						},
					},
				},
			},
		}
		conditions := ParentConditions{
			makeParentRef("ns-b", "gw-b", "listener-1"): {
				{Type: string(gatewayapi.RouteConditionAccepted), Status: kube_meta.ConditionTrue, Reason: "Accepted"},
			},
		}

		mergeHTTPRouteStatus(route, conditions)

		Expect(route.Status.Parents).To(HaveLen(3))
		// Other controller statuses come first (preserved order), then ours
		Expect(route.Status.Parents[0].ControllerName).To(Equal(otherController))
		Expect(route.Status.Parents[1].ControllerName).To(Equal(otherController))
		Expect(route.Status.Parents[2].ControllerName).To(Equal(common.ControllerName))
	})
})
