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
	gatewayapi_v1 "sigs.k8s.io/gateway-api/apis/v1"
	gatewayapi "sigs.k8s.io/gateway-api/apis/v1beta1"

	"github.com/kumahq/kuma/pkg/plugins/bootstrap/k8s/scheme"
	"github.com/kumahq/kuma/pkg/plugins/runtime/k8s/controllers/gatewayapi/attachment"
	"github.com/kumahq/kuma/pkg/plugins/runtime/k8s/controllers/gatewayapi/common"
)

var k8sScheme *kube_runtime.Scheme

var _ = BeforeSuite(func() {
	var err error
	k8sScheme, err = scheme.NewScheme()
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

	checkAttachment := func(expected attachment.Attachment, expectedKind attachment.Kind) func([]gatewayapi.Hostname) {
		return func(routeHostnames []gatewayapi.Hostname) {
			res, kind, err := attachment.EvaluateParentRefAttachment(
				context.Background(),
				kubeClient,
				routeHostnames,
				defaultRouteNs,
				simpleRef,
			)
			Expect(err).ToNot(HaveOccurred())
			Expect(res).To(Equal(expected))
			Expect(kind).To(Equal(expectedKind))
		}
	}

	Context("all listeners", func() {
		BeforeEach(func() {
			simpleRef = *gatewayRef.DeepCopy()
			simpleRef.Name = gatewayapi.ObjectName(gatewayMultipleListeners.Name)
		})

		DescribeTable("matches some listener", checkAttachment(attachment.Allowed, attachment.Gateway),
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

		DescribeTable("matches", checkAttachment(attachment.Allowed, attachment.Gateway),
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

		DescribeTable("matches", checkAttachment(attachment.Allowed, attachment.Gateway),
			Entry("on exact match", []gatewayapi.Hostname{"simple.local"}),
			Entry("if one matches", []gatewayapi.Hostname{"other.local", "simple.local"}),
			Entry("if route wildcard matches", []gatewayapi.Hostname{"*.local"}),
		)

		DescribeTable("doesn't match", checkAttachment(attachment.NoHostnameIntersection, attachment.Gateway),
			Entry("without intersection", []gatewayapi.Hostname{"other.local"}),
		)
	})

	Context("listener with wildcard hostname", func() {
		BeforeEach(func() {
			simpleRef = *gatewayRef.DeepCopy()
			simpleRef.Name = gatewayapi.ObjectName(gatewayMultipleListeners.Name)
			simpleRef.SectionName = &wildcardListenerName
		})

		DescribeTable("matches", checkAttachment(attachment.Allowed, attachment.Gateway),
			Entry("with a complete hostname", []gatewayapi.Hostname{"something.wildcard.local"}),
			Entry("with a complete hostname with extra subdomains", []gatewayapi.Hostname{"something.else.wildcard.local"}),
			Entry("with a complete hostname with extra subdomain and wildcard", []gatewayapi.Hostname{"*.else.wildcard.local"}),
			Entry("with a wildcard hostname", []gatewayapi.Hostname{"*.wildcard.local"}),
		)

		DescribeTable("doesn't match", checkAttachment(attachment.NoHostnameIntersection, attachment.Gateway),
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
			res, kind, err := attachment.EvaluateParentRefAttachment(
				context.Background(),
				kubeClient,
				nil,
				defaultRouteNs,
				simpleRef,
			)
			Expect(err).ToNot(HaveOccurred())
			Expect(res).To(Equal(attachment.Allowed))
			Expect(kind).To(Equal(attachment.Gateway))
		})
		It("allows from same namespace for all listeners", func() {
			simpleRef.SectionName = nil
			res, kind, err := attachment.EvaluateParentRefAttachment(
				context.Background(),
				kubeClient,
				nil,
				defaultRouteNs,
				simpleRef,
			)
			Expect(err).ToNot(HaveOccurred())
			Expect(res).To(Equal(attachment.Allowed))
			Expect(kind).To(Equal(attachment.Gateway))
		})
		It("denies from other namespace", func() {
			simpleRef.SectionName = &simpleListenerName
			ns := gatewayapi.Namespace(defaultNs)
			simpleRef.Namespace = &ns

			res, kind, err := attachment.EvaluateParentRefAttachment(
				context.Background(),
				kubeClient,
				nil,
				otherRouteNs,
				simpleRef,
			)
			Expect(err).ToNot(HaveOccurred())
			Expect(res).To(Equal(attachment.NotPermitted))
			Expect(kind).To(Equal(attachment.Gateway))
		})
		It("denies from other namespace for all listeners", func() {
			simpleRef.SectionName = nil
			ns := gatewayapi.Namespace(defaultNs)
			simpleRef.Namespace = &ns

			res, kind, err := attachment.EvaluateParentRefAttachment(
				context.Background(),
				kubeClient,
				nil,
				otherRouteNs,
				simpleRef,
			)
			Expect(err).ToNot(HaveOccurred())
			Expect(res).To(Equal(attachment.NotPermitted))
			Expect(kind).To(Equal(attachment.Gateway))
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

			res, kind, err := attachment.EvaluateParentRefAttachment(
				context.Background(),
				kubeClient,
				nil,
				defaultRouteNs,
				parentRef,
			)
			Expect(err).ToNot(HaveOccurred())
			Expect(res).To(Equal(attachment.Allowed))
			Expect(kind).To(Equal(attachment.Gateway))
		})
		It("allows route from other NS", func() {
			parentRef.SectionName = &allNsListenerName
			ns := gatewayapi.Namespace(defaultNs)
			parentRef.Namespace = &ns

			res, kind, err := attachment.EvaluateParentRefAttachment(
				context.Background(),
				kubeClient,
				nil,
				otherRouteNs,
				parentRef,
			)
			Expect(err).ToNot(HaveOccurred())
			Expect(res).To(Equal(attachment.Allowed))
			Expect(kind).To(Equal(attachment.Gateway))
		})
		It("allows from other NS for all listeners", func() {
			ns := gatewayapi.Namespace(defaultNs)
			parentRef.Namespace = &ns

			res, kind, err := attachment.EvaluateParentRefAttachment(
				context.Background(),
				kubeClient,
				nil,
				otherRouteNs,
				parentRef,
			)
			Expect(err).ToNot(HaveOccurred())
			Expect(res).To(Equal(attachment.Allowed))
			Expect(kind).To(Equal(attachment.Gateway))
		})
	})
})

var _ = Describe("NoMatchingParent support", func() {
	var simpleRef gatewayapi.ParentReference

	var kubeClient kube_client.Client
	BeforeEach(func() {
		kubeClient = kube_client_fake.NewClientBuilder().WithScheme(k8sScheme).WithObjects(
			gatewayClass,
			gateway,
			gatewayMultipleListeners,
			defaultRouteNs,
			otherRouteNs,
		).Build()
		simpleRef = *gatewayRef.DeepCopy()
		simpleRef.Name = gatewayapi.ObjectName(gateway.Name)
	})

	It("allows no SectionName", func() {
		res, kind, err := attachment.EvaluateParentRefAttachment(
			context.Background(),
			kubeClient,
			nil,
			defaultRouteNs,
			simpleRef,
		)
		Expect(err).ToNot(HaveOccurred())
		Expect(res).To(Equal(attachment.Allowed))
		Expect(kind).To(Equal(attachment.Gateway))
	})

	It("allows matching SectionName", func() {
		simpleRef.SectionName = &simpleListenerName
		res, kind, err := attachment.EvaluateParentRefAttachment(
			context.Background(),
			kubeClient,
			nil,
			defaultRouteNs,
			simpleRef,
		)
		Expect(err).ToNot(HaveOccurred())
		Expect(res).To(Equal(attachment.Allowed))
		Expect(kind).To(Equal(attachment.Gateway))
	})

	It("doesn't allow SectionName mismatches", func() {
		someNonexistentSection := gatewayapi.SectionName("someNonexistentSection")
		simpleRef.SectionName = &someNonexistentSection
		res, kind, err := attachment.EvaluateParentRefAttachment(
			context.Background(),
			kubeClient,
			nil,
			defaultRouteNs,
			simpleRef,
		)
		Expect(err).ToNot(HaveOccurred())
		Expect(res).To(Equal(attachment.NoMatchingParent))
		Expect(kind).To(Equal(attachment.Gateway))
	})

	It("allows Port matches", func() {
		simpleRef.Port = &simpleListener.Port
		res, kind, err := attachment.EvaluateParentRefAttachment(
			context.Background(),
			kubeClient,
			nil,
			defaultRouteNs,
			simpleRef,
		)
		Expect(err).ToNot(HaveOccurred())
		Expect(res).To(Equal(attachment.Allowed))
		Expect(kind).To(Equal(attachment.Gateway))
	})

	It("allows SectionName & Port matches", func() {
		simpleRef.SectionName = &simpleListenerName
		simpleRef.Port = &simpleListener.Port
		res, kind, err := attachment.EvaluateParentRefAttachment(
			context.Background(),
			kubeClient,
			nil,
			defaultRouteNs,
			simpleRef,
		)
		Expect(err).ToNot(HaveOccurred())
		Expect(res).To(Equal(attachment.Allowed))
		Expect(kind).To(Equal(attachment.Gateway))
	})

	It("doesn't allow SectionName mismatches but Port matches", func() {
		someNonexistentSection := gatewayapi.SectionName("someNonexistentSection")
		simpleRef.SectionName = &someNonexistentSection
		simpleRef.Port = &simpleListener.Port
		res, kind, err := attachment.EvaluateParentRefAttachment(
			context.Background(),
			kubeClient,
			nil,
			defaultRouteNs,
			simpleRef,
		)
		Expect(err).ToNot(HaveOccurred())
		Expect(res).To(Equal(attachment.NoMatchingParent))
		Expect(kind).To(Equal(attachment.Gateway))
	})

	It("doesn't allow SectionName matches but Port mismatches", func() {
		simpleRef.SectionName = &simpleListenerName
		someNonexistentPort := gatewayapi.PortNumber(10101)
		simpleRef.Port = &someNonexistentPort
		res, kind, err := attachment.EvaluateParentRefAttachment(
			context.Background(),
			kubeClient,
			nil,
			defaultRouteNs,
			simpleRef,
		)
		Expect(err).ToNot(HaveOccurred())
		Expect(res).To(Equal(attachment.NoMatchingParent))
		Expect(kind).To(Equal(attachment.Gateway))
	})
})

var (
	defaultNs       = "default"
	otherNs         = "other"
	fromAll         = gatewayapi_v1.NamespacesFromAll
	fromSame        = gatewayapi_v1.NamespacesFromSame
	gatewayGroup    = gatewayapi.Group(gatewayapi.GroupName)
	gatewayKind     = gatewayapi.Kind("Gateway")
	simpleHostname  = gatewayapi.Hostname("simple.local")
	anyTestHostname = gatewayapi.Hostname("*.wildcard.local")

	listenerProgrammed = []kube_meta.Condition{
		{
			Type:   string(gatewayapi_v1.ListenerConditionProgrammed),
			Status: kube_meta.ConditionTrue,
		},
	}

	simpleListenerName      = gatewayapi.SectionName("simple")
	anyHostnameListenerName = simpleListenerName
	simpleListener          = gatewayapi.Listener{
		Name:     simpleListenerName,
		Port:     gatewayapi.PortNumber(80),
		Protocol: gatewayapi_v1.HTTPProtocolType,
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
		Protocol: gatewayapi_v1.HTTPProtocolType,
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
		Protocol: gatewayapi_v1.HTTPProtocolType,
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
					Conditions: listenerProgrammed,
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
					Conditions: listenerProgrammed,
				},
				{
					Name:       wildcardListenerName,
					Conditions: listenerProgrammed,
				},
				{
					Name:       allNsListenerName,
					Conditions: listenerProgrammed,
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
