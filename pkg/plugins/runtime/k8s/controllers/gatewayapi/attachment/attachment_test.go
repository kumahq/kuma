package attachment_test

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	kube_core "k8s.io/api/core/v1"
	kube_meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	kube_runtime "k8s.io/apimachinery/pkg/runtime"
	kube_client "sigs.k8s.io/controller-runtime/pkg/client"
	kube_client_fake "sigs.k8s.io/controller-runtime/pkg/client/fake"
	gatewayapi "sigs.k8s.io/gateway-api/apis/v1beta1"

	"github.com/kumahq/kuma/pkg/plugins/bootstrap/k8s"
	"github.com/kumahq/kuma/pkg/plugins/runtime/k8s/controllers/gatewayapi/attachment"
	"github.com/kumahq/kuma/pkg/plugins/runtime/k8s/controllers/gatewayapi/common"
)

var k8sScheme *kube_runtime.Scheme

var _ = BeforeSuite(func() {
	var err error
	k8sScheme, err = k8s.NewScheme()
	Expect(err).NotTo(HaveOccurred())
})

var _ = Describe("Hostname intersection support", func() {
	var kubeClient kube_client.Client
	var simpleRef gatewayapi.ParentReference
	BeforeEach(func() {
		kubeClient = kube_client_fake.NewClientBuilder().WithScheme(k8sScheme).WithObjects(
			gatewayClass,
			gateway,
			gatewayMultipleListeners,
			defaultRouteNs,
			otherRouteNs,
		).Build()
	})

	checkAttachment := func(expected attachment.Attachment) func([]gatewayapi.Hostname) {
		return func(routeHostnames []gatewayapi.Hostname) {
			res, err := attachment.EvaluateParentRefAttachment(
				context.Background(),
				kubeClient,
				routeHostnames,
				defaultRouteNs,
				simpleRef,
			)
			Expect(err).ToNot(HaveOccurred())
			Expect(res).To(Equal(expected))
		}
	}

	Context("all listeners", func() {
		BeforeEach(func() {
			simpleRef = *gatewayRef.DeepCopy()
			simpleRef.Name = gatewayapi.ObjectName(gatewayMultipleListeners.Name)
		})

		DescribeTable("matches some listener", checkAttachment(attachment.Allowed),
			Entry("exact match", []gatewayapi.Hostname{"other.local"}),
			Entry("non matching wildcard", []gatewayapi.Hostname{"*.wildcard.io", "something.else.local"}),
		)
	})

	Context("listener without hostname", func() {
		BeforeEach(func() {
			simpleRef = *gatewayRef.DeepCopy()
			simpleRef.Name = gatewayapi.ObjectName(gatewayMultipleListeners.Name)
			simpleRef.SectionName = &anyHostnameListenerName
		})

		DescribeTable("matches", checkAttachment(attachment.Allowed),
			Entry("without route hostname", nil),
			Entry("with some hostname", []gatewayapi.Hostname{"other.local"}),
		)
	})

	Context("listener with simple hostname", func() {
		BeforeEach(func() {
			simpleRef = *gatewayRef.DeepCopy()
			simpleRef.Name = gatewayapi.ObjectName(gatewayMultipleListeners.Name)
			simpleRef.SectionName = &simpleHostnameListenerName
		})

		DescribeTable("matches", checkAttachment(attachment.Allowed),
			Entry("on exact match", []gatewayapi.Hostname{"simple.local"}),
			Entry("if one matches", []gatewayapi.Hostname{"other.local", "simple.local"}),
			Entry("if route wildcard matches", []gatewayapi.Hostname{"*.local"}),
		)

		DescribeTable("doesn't match", checkAttachment(attachment.NoHostnameIntersection),
			Entry("without intersection", []gatewayapi.Hostname{"other.local"}),
		)
	})

	Context("listener with wildcard hostname", func() {
		BeforeEach(func() {
			simpleRef = *gatewayRef.DeepCopy()
			simpleRef.Name = gatewayapi.ObjectName(gatewayMultipleListeners.Name)
			simpleRef.SectionName = &wildcardListenerName
		})

		DescribeTable("matches", checkAttachment(attachment.Allowed),
			Entry("with a complete hostname", []gatewayapi.Hostname{"something.wildcard.local"}),
			Entry("with a complete hostname with extra subdomains", []gatewayapi.Hostname{"something.else.wildcard.local"}),
			Entry("with a complete hostname with extra subdomain and wildcard", []gatewayapi.Hostname{"*.else.wildcard.local"}),
			Entry("with a wildcard hostname", []gatewayapi.Hostname{"*.wildcard.local"}),
		)

		DescribeTable("doesn't match", checkAttachment(attachment.NoHostnameIntersection),
			Entry("wildcard suffix without subdomain", []gatewayapi.Hostname{"wildcard.local"}),
			Entry("without intersection", []gatewayapi.Hostname{"other.local"}),
		)
	})
})

var _ = Describe("AllowedRoutes support", func() {
	var kubeClient kube_client.Client
	BeforeEach(func() {
		kubeClient = kube_client_fake.NewClientBuilder().WithScheme(k8sScheme).WithObjects(
			gatewayClass,
			gateway,
			gatewayMultipleListeners,
			defaultRouteNs,
			otherRouteNs,
		).Build()

	})
	Context("default AllowedRoutes", func() {
		var simpleRef gatewayapi.ParentReference
		BeforeEach(func() {
			simpleRef = *gatewayRef.DeepCopy()
			simpleRef.Name = gatewayapi.ObjectName(gateway.Name)
		})

		It("allows from same namespace", func() {
			simpleRef.SectionName = &simpleListenerName
			res, err := attachment.EvaluateParentRefAttachment(
				context.Background(),
				kubeClient,
				nil,
				defaultRouteNs,
				simpleRef,
			)
			Expect(err).ToNot(HaveOccurred())
			Expect(res).To(Equal(attachment.Allowed))
		})
		It("allows from same namespace for all listeners", func() {
			simpleRef.SectionName = nil
			res, err := attachment.EvaluateParentRefAttachment(
				context.Background(),
				kubeClient,
				nil,
				defaultRouteNs,
				simpleRef,
			)
			Expect(err).ToNot(HaveOccurred())
			Expect(res).To(Equal(attachment.Allowed))
		})
		It("denies from other namespace", func() {
			simpleRef.SectionName = &simpleListenerName
			ns := gatewayapi.Namespace(defaultNs)
			simpleRef.Namespace = &ns

			res, err := attachment.EvaluateParentRefAttachment(
				context.Background(),
				kubeClient,
				nil,
				otherRouteNs,
				simpleRef,
			)
			Expect(err).ToNot(HaveOccurred())
			Expect(res).To(Equal(attachment.NotPermitted))
		})
		It("denies from other namespace for all listeners", func() {
			simpleRef.SectionName = nil
			ns := gatewayapi.Namespace(defaultNs)
			simpleRef.Namespace = &ns

			res, err := attachment.EvaluateParentRefAttachment(
				context.Background(),
				kubeClient,
				nil,
				otherRouteNs,
				simpleRef,
			)
			Expect(err).ToNot(HaveOccurred())
			Expect(res).To(Equal(attachment.NotPermitted))
		})
	})
	Context("FromAll", func() {
		var parentRef gatewayapi.ParentReference
		BeforeEach(func() {
			parentRef = *gatewayRef.DeepCopy()
			parentRef.Name = gatewayapi.ObjectName(gatewayMultipleListeners.Name)
		})

		It("allows route from same NS", func() {
			parentRef.SectionName = &allNsListenerName

			res, err := attachment.EvaluateParentRefAttachment(
				context.Background(),
				kubeClient,
				nil,
				defaultRouteNs,
				parentRef,
			)
			Expect(err).ToNot(HaveOccurred())
			Expect(res).To(Equal(attachment.Allowed))
		})
		It("allows route from other NS", func() {
			parentRef.SectionName = &allNsListenerName
			ns := gatewayapi.Namespace(defaultNs)
			parentRef.Namespace = &ns

			res, err := attachment.EvaluateParentRefAttachment(
				context.Background(),
				kubeClient,
				nil,
				otherRouteNs,
				parentRef,
			)
			Expect(err).ToNot(HaveOccurred())
			Expect(res).To(Equal(attachment.Allowed))
		})
		It("allows from other NS for all listeners", func() {
			ns := gatewayapi.Namespace(defaultNs)
			parentRef.Namespace = &ns

			res, err := attachment.EvaluateParentRefAttachment(
				context.Background(),
				kubeClient,
				nil,
				otherRouteNs,
				parentRef,
			)
			Expect(err).ToNot(HaveOccurred())
			Expect(res).To(Equal(attachment.Allowed))
		})
	})
})

var (
	defaultNs       = "default"
	otherNs         = "other"
	fromAll         = gatewayapi.NamespacesFromAll
	fromSame        = gatewayapi.NamespacesFromSame
	gatewayGroup    = gatewayapi.Group(gatewayapi.GroupName)
	gatewayKind     = gatewayapi.Kind("Gateway")
	simpleHostname  = gatewayapi.Hostname("simple.local")
	anyTestHostname = gatewayapi.Hostname("*.wildcard.local")

	listenerReady = []kube_meta.Condition{
		{
			Type:   string(gatewayapi.ListenerConditionReady),
			Status: kube_meta.ConditionTrue,
		},
	}

	simpleListenerName      = gatewayapi.SectionName("simple")
	anyHostnameListenerName = simpleListenerName
	simpleListener          = gatewayapi.Listener{
		Name:     simpleListenerName,
		Port:     gatewayapi.PortNumber(80),
		Protocol: gatewayapi.HTTPProtocolType,
		AllowedRoutes: &gatewayapi.AllowedRoutes{
			Namespaces: &gatewayapi.RouteNamespaces{
				From: &fromSame,
			},
		},
	}
	wildcardListenerName = gatewayapi.SectionName("wildcard")
	wildcardListener     = gatewayapi.Listener{
		Name:     wildcardListenerName,
		Port:     gatewayapi.PortNumber(80),
		Protocol: gatewayapi.HTTPProtocolType,
		Hostname: &anyTestHostname,
		AllowedRoutes: &gatewayapi.AllowedRoutes{
			Namespaces: &gatewayapi.RouteNamespaces{
				From: &fromSame,
			},
		},
	}
	allNsListenerName          = gatewayapi.SectionName("allNS")
	simpleHostnameListenerName = allNsListenerName
	allNsListener              = gatewayapi.Listener{
		Name:     allNsListenerName,
		Port:     gatewayapi.PortNumber(80),
		Protocol: gatewayapi.HTTPProtocolType,
		Hostname: &simpleHostname,
		AllowedRoutes: &gatewayapi.AllowedRoutes{
			Namespaces: &gatewayapi.RouteNamespaces{
				From: &fromAll,
			},
		},
	}
	gatewayClass = &gatewayapi.GatewayClass{
		ObjectMeta: kube_meta.ObjectMeta{
			Name: "kuma",
		},
		Spec: gatewayapi.GatewayClassSpec{
			ControllerName: common.ControllerName,
		},
	}
	gateway = &gatewayapi.Gateway{
		ObjectMeta: kube_meta.ObjectMeta{
			Name:      "gateway",
			Namespace: defaultNs,
		},
		Spec: gatewayapi.GatewaySpec{
			GatewayClassName: "kuma",
			Listeners: []gatewayapi.Listener{
				simpleListener,
			},
		},
		Status: gatewayapi.GatewayStatus{
			Listeners: []gatewayapi.ListenerStatus{
				{
					Name:       simpleListenerName,
					Conditions: listenerReady,
				},
			},
		},
	}
	gatewayRef = gatewayapi.ParentReference{
		Group: &gatewayGroup,
		Kind:  &gatewayKind,
	}

	gatewayMultipleListeners = &gatewayapi.Gateway{
		ObjectMeta: kube_meta.ObjectMeta{
			Name:      "multigateway",
			Namespace: defaultNs,
		},
		Spec: gatewayapi.GatewaySpec{
			GatewayClassName: "kuma",
			Listeners: []gatewayapi.Listener{
				wildcardListener,
				allNsListener,
				simpleListener,
			},
		},
		Status: gatewayapi.GatewayStatus{
			Listeners: []gatewayapi.ListenerStatus{
				{
					Name:       simpleListenerName,
					Conditions: listenerReady,
				},
				{
					Name:       wildcardListenerName,
					Conditions: listenerReady,
				},
				{
					Name:       allNsListenerName,
					Conditions: listenerReady,
				},
			},
		},
	}

	defaultRouteNs = &kube_core.Namespace{
		ObjectMeta: kube_meta.ObjectMeta{
			Name: defaultNs,
		},
	}
	otherRouteNs = &kube_core.Namespace{
		ObjectMeta: kube_meta.ObjectMeta{
			Name: otherNs,
		},
	}
)
