package gatewayapi

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	kube_meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	gatewayapi "sigs.k8s.io/gateway-api/apis/v1beta1"

	"github.com/kumahq/kuma/v3/pkg/plugins/runtime/k8s/controllers/gatewayapi/common"
	"github.com/kumahq/kuma/v3/pkg/util/pointer"
)

var _ = Describe("mergeHTTPRouteStatus", func() {
	makeParentRef := func(namespace, name, sectionName string) gatewayapi.ParentReference {
		group := gatewayapi.Group(gatewayapi.GroupVersion.Group)
		kind := gatewayapi.Kind("Service")
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
			makeParentRef("ns-b", "svc-z", "listener-1"): {
				{Type: string(gatewayapi.RouteConditionAccepted), Status: kube_meta.ConditionTrue, Reason: "Accepted"},
			},
			makeParentRef("ns-a", "svc-a", "listener-1"): {
				{Type: string(gatewayapi.RouteConditionAccepted), Status: kube_meta.ConditionTrue, Reason: "Accepted"},
			},
			makeParentRef("ns-a", "svc-a", "listener-2"): {
				{Type: string(gatewayapi.RouteConditionAccepted), Status: kube_meta.ConditionTrue, Reason: "Accepted"},
			},
		}

		mergeHTTPRouteStatus(route, conditions)

		Expect(route.Status.Parents).To(HaveLen(3))
		Expect(route.Status.Parents[0].ParentRef.Name).To(Equal(gatewayapi.ObjectName("svc-a")))
		Expect(pointer.Deref(route.Status.Parents[0].ParentRef.SectionName)).To(Equal(gatewayapi.SectionName("listener-1")))
		Expect(route.Status.Parents[1].ParentRef.Name).To(Equal(gatewayapi.ObjectName("svc-a")))
		Expect(pointer.Deref(route.Status.Parents[1].ParentRef.SectionName)).To(Equal(gatewayapi.SectionName("listener-2")))
		Expect(route.Status.Parents[2].ParentRef.Name).To(Equal(gatewayapi.ObjectName("svc-z")))
	})

	It("preserves statuses from other controllers unsorted", func() {
		otherController := gatewayapi.GatewayController("other-controller")
		route := &gatewayapi.HTTPRoute{
			Status: gatewayapi.HTTPRouteStatus{
				RouteStatus: gatewayapi.RouteStatus{
					Parents: []gatewayapi.RouteParentStatus{
						{
							ParentRef:      makeParentRef("ns-z", "svc-z", "listener-1"),
							ControllerName: otherController,
						},
						{
							ParentRef:      makeParentRef("ns-a", "svc-a", "listener-1"),
							ControllerName: otherController,
						},
					},
				},
			},
		}
		conditions := ParentConditions{
			makeParentRef("ns-b", "svc-b", "listener-1"): {
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

	It("preserves previous status when parent conditions are intentionally empty", func() {
		ref := makeParentRef("ns-a", "gateway-a", "listener-1")
		route := &gatewayapi.HTTPRoute{
			Status: gatewayapi.HTTPRouteStatus{
				RouteStatus: gatewayapi.RouteStatus{
					Parents: []gatewayapi.RouteParentStatus{
						{
							ParentRef:      ref,
							ControllerName: common.ControllerName,
							Conditions: []kube_meta.Condition{
								{Type: string(gatewayapi.RouteConditionAccepted), Status: kube_meta.ConditionTrue, Reason: "Accepted"},
							},
						},
					},
				},
			},
		}

		mergeHTTPRouteStatus(route, ParentConditions{ref: nil})

		Expect(route.Status.Parents).To(HaveLen(1))
		Expect(route.Status.Parents[0].ParentRef).To(Equal(ref))
		Expect(route.Status.Parents[0].Conditions).To(HaveLen(1))
		Expect(route.Status.Parents[0].Conditions[0].Reason).To(Equal("Accepted"))
	})

	It("does not create an empty status when parent conditions are intentionally empty", func() {
		ref := makeParentRef("ns-a", "gateway-a", "listener-1")
		route := &gatewayapi.HTTPRoute{}

		mergeHTTPRouteStatus(route, ParentConditions{ref: nil})

		Expect(route.Status.Parents).To(BeEmpty())
	})
})
