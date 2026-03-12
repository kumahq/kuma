package controllers_test

import (
	"context"

	"github.com/go-logr/logr"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	kube_core "k8s.io/api/core/v1"
	kube_discovery "k8s.io/api/discovery/v1"
	kube_meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	kube_types "k8s.io/apimachinery/pkg/types"
	kube_events "k8s.io/client-go/tools/events"
	kube_ctrl "sigs.k8s.io/controller-runtime"
	kube_client "sigs.k8s.io/controller-runtime/pkg/client"
	kube_client_fake "sigs.k8s.io/controller-runtime/pkg/client/fake"

	mesh_proto "github.com/kumahq/kuma/v2/api/mesh/v1alpha1"
	meshzoneaddress_k8s "github.com/kumahq/kuma/v2/pkg/core/resources/apis/meshzoneaddress/k8s/v1alpha1"
	. "github.com/kumahq/kuma/v2/pkg/plugins/runtime/k8s/controllers"
	"github.com/kumahq/kuma/v2/pkg/plugins/runtime/k8s/metadata"
)

const (
	testZone      = "zone-1"
	testNamespace = "kuma-system"
	testSvcName   = "zone-ingress"
)

func boolPtr(b bool) *bool { return &b }

func newNamespace(meshLabel string) *kube_core.Namespace {
	ns := &kube_core.Namespace{
		ObjectMeta: kube_meta.ObjectMeta{Name: testNamespace},
	}
	if meshLabel != "" {
		ns.Labels = map[string]string{mesh_proto.MeshTag: meshLabel}
	}
	return ns
}

func newZoneProxyService(svcType kube_core.ServiceType, extraLabels map[string]string) *kube_core.Service {
	labels := map[string]string{
		metadata.KumaZoneProxyTypeLabel: metadata.KumaZoneProxyTypeIngress,
	}
	for k, v := range extraLabels {
		labels[k] = v
	}
	return &kube_core.Service{
		ObjectMeta: kube_meta.ObjectMeta{
			Name:      testSvcName,
			Namespace: testNamespace,
			Labels:    labels,
		},
		Spec: kube_core.ServiceSpec{Type: svcType},
	}
}

func newReadyEndpointSlice() *kube_discovery.EndpointSlice {
	return &kube_discovery.EndpointSlice{
		ObjectMeta: kube_meta.ObjectMeta{
			Name:      testSvcName + "-abc",
			Namespace: testNamespace,
			Labels:    map[string]string{kube_discovery.LabelServiceName: testSvcName},
		},
		AddressType: kube_discovery.AddressTypeIPv4,
		Endpoints: []kube_discovery.Endpoint{
			{Addresses: []string{"10.0.0.1"}, Conditions: kube_discovery.EndpointConditions{Ready: boolPtr(true)}},
		},
	}
}

func reconcileMZA(kubeClient kube_client.Client, recorder kube_events.EventRecorder) error {
	reconciler := &MeshZoneAddressReconciler{
		Client:        kubeClient,
		Log:           logr.Discard(),
		Scheme:        k8sClientScheme,
		EventRecorder: recorder,
		ZoneName:      testZone,
	}
	_, err := reconciler.Reconcile(context.Background(), kube_ctrl.Request{
		NamespacedName: kube_types.NamespacedName{Name: testSvcName, Namespace: testNamespace},
	})
	return err
}

func listMZAs(kubeClient kube_client.Client) *meshzoneaddress_k8s.MeshZoneAddressList {
	mzas := &meshzoneaddress_k8s.MeshZoneAddressList{}
	Expect(kubeClient.List(context.Background(), mzas)).To(Succeed())
	return mzas
}

var _ = Describe("MeshZoneAddressReconciler", func() {
	var recorder *kube_events.FakeRecorder

	BeforeEach(func() {
		recorder = kube_events.NewFakeRecorder(10)
	})

	It("skips service without zone-proxy label", func() {
		svc := &kube_core.Service{
			ObjectMeta: kube_meta.ObjectMeta{Name: testSvcName, Namespace: testNamespace},
			Spec: kube_core.ServiceSpec{
				Type:  kube_core.ServiceTypeLoadBalancer,
				Ports: []kube_core.ServicePort{{Port: 10001}},
			},
			Status: kube_core.ServiceStatus{
				LoadBalancer: kube_core.LoadBalancerStatus{
					Ingress: []kube_core.LoadBalancerIngress{{Hostname: "lb.example.com"}},
				},
			},
		}
		kubeClient := kube_client_fake.NewClientBuilder().
			WithScheme(k8sClientScheme).
			WithObjects(newNamespace(""), svc, newReadyEndpointSlice()).
			Build()

		Expect(reconcileMZA(kubeClient, recorder)).To(Succeed())
		Expect(listMZAs(kubeClient).Items).To(BeEmpty())
	})

	It("skips when no ready endpoints", func() {
		svc := newZoneProxyService(kube_core.ServiceTypeLoadBalancer, nil)
		svc.Spec.Ports = []kube_core.ServicePort{{Port: 10001}}
		svc.Status = kube_core.ServiceStatus{
			LoadBalancer: kube_core.LoadBalancerStatus{
				Ingress: []kube_core.LoadBalancerIngress{{Hostname: "lb.example.com"}},
			},
		}
		notReady := &kube_discovery.EndpointSlice{
			ObjectMeta: kube_meta.ObjectMeta{
				Name:      testSvcName + "-abc",
				Namespace: testNamespace,
				Labels:    map[string]string{kube_discovery.LabelServiceName: testSvcName},
			},
			AddressType: kube_discovery.AddressTypeIPv4,
			Endpoints: []kube_discovery.Endpoint{
				{Addresses: []string{"10.0.0.1"}, Conditions: kube_discovery.EndpointConditions{Ready: boolPtr(false)}},
			},
		}
		kubeClient := kube_client_fake.NewClientBuilder().
			WithScheme(k8sClientScheme).
			WithObjects(newNamespace(""), svc, notReady).
			Build()

		Expect(reconcileMZA(kubeClient, recorder)).To(Succeed())
		Expect(listMZAs(kubeClient).Items).To(BeEmpty())
	})

	It("creates MeshZoneAddress from LoadBalancer hostname (hostname > IP)", func() {
		svc := newZoneProxyService(kube_core.ServiceTypeLoadBalancer, map[string]string{
			mesh_proto.MeshTag: "demo",
		})
		svc.Spec.Ports = []kube_core.ServicePort{{Port: 10001}}
		svc.Status = kube_core.ServiceStatus{
			LoadBalancer: kube_core.LoadBalancerStatus{
				Ingress: []kube_core.LoadBalancerIngress{{Hostname: "lb.example.com", IP: "1.2.3.4"}},
			},
		}
		kubeClient := kube_client_fake.NewClientBuilder().
			WithScheme(k8sClientScheme).
			WithObjects(newNamespace(""), svc, newReadyEndpointSlice()).
			Build()

		Expect(reconcileMZA(kubeClient, recorder)).To(Succeed())

		mzas := listMZAs(kubeClient)
		Expect(mzas.Items).To(HaveLen(1))
		mza := mzas.Items[0]
		Expect(mza.Name).To(Equal(testSvcName))
		Expect(mza.Namespace).To(Equal(testNamespace))
		Expect(mza.Labels[mesh_proto.MeshTag]).To(Equal("demo"))
		Expect(mza.Labels[mesh_proto.ZoneTag]).To(Equal(testZone))
		Expect(mza.Labels[mesh_proto.ManagedByLabel]).To(Equal("k8s-controller"))
		Expect(mza.Spec.Address).To(Equal("lb.example.com"))
		Expect(mza.Spec.Port).To(Equal(int32(10001)))
	})

	It("creates MeshZoneAddress from LoadBalancer IP when no hostname", func() {
		svc := newZoneProxyService(kube_core.ServiceTypeLoadBalancer, nil)
		svc.Spec.Ports = []kube_core.ServicePort{{Port: 10001}}
		svc.Status = kube_core.ServiceStatus{
			LoadBalancer: kube_core.LoadBalancerStatus{
				Ingress: []kube_core.LoadBalancerIngress{{IP: "5.6.7.8"}},
			},
		}
		kubeClient := kube_client_fake.NewClientBuilder().
			WithScheme(k8sClientScheme).
			WithObjects(newNamespace(""), svc, newReadyEndpointSlice()).
			Build()

		Expect(reconcileMZA(kubeClient, recorder)).To(Succeed())

		mzas := listMZAs(kubeClient)
		Expect(mzas.Items).To(HaveLen(1))
		Expect(mzas.Items[0].Spec.Address).To(Equal("5.6.7.8"))
		Expect(mzas.Items[0].Spec.Port).To(Equal(int32(10001)))
	})

	It("skips LoadBalancer not yet provisioned (no ingress)", func() {
		svc := newZoneProxyService(kube_core.ServiceTypeLoadBalancer, nil)
		svc.Spec.Ports = []kube_core.ServicePort{{Port: 10001}}
		// Status.LoadBalancer.Ingress left empty
		kubeClient := kube_client_fake.NewClientBuilder().
			WithScheme(k8sClientScheme).
			WithObjects(newNamespace(""), svc, newReadyEndpointSlice()).
			Build()

		Expect(reconcileMZA(kubeClient, recorder)).To(Succeed())
		Expect(listMZAs(kubeClient).Items).To(BeEmpty())
	})

	It("creates MeshZoneAddress from NodePort using node ExternalIP", func() {
		svc := newZoneProxyService(kube_core.ServiceTypeNodePort, nil)
		svc.Spec.Ports = []kube_core.ServicePort{{Port: 10001, NodePort: 30001}}
		node := &kube_core.Node{
			ObjectMeta: kube_meta.ObjectMeta{Name: "node-1"},
			Status: kube_core.NodeStatus{
				Addresses: []kube_core.NodeAddress{
					{Type: kube_core.NodeInternalIP, Address: "192.168.1.1"},
					{Type: kube_core.NodeExternalIP, Address: "1.2.3.4"},
				},
			},
		}
		kubeClient := kube_client_fake.NewClientBuilder().
			WithScheme(k8sClientScheme).
			WithObjects(newNamespace(""), svc, newReadyEndpointSlice(), node).
			Build()

		Expect(reconcileMZA(kubeClient, recorder)).To(Succeed())

		mzas := listMZAs(kubeClient)
		Expect(mzas.Items).To(HaveLen(1))
		Expect(mzas.Items[0].Spec.Address).To(Equal("1.2.3.4"))
		Expect(mzas.Items[0].Spec.Port).To(Equal(int32(30001)))
	})

	It("creates MeshZoneAddress from NodePort using node InternalIP when no ExternalIP", func() {
		svc := newZoneProxyService(kube_core.ServiceTypeNodePort, nil)
		svc.Spec.Ports = []kube_core.ServicePort{{Port: 10001, NodePort: 30001}}
		node := &kube_core.Node{
			ObjectMeta: kube_meta.ObjectMeta{Name: "node-1"},
			Status: kube_core.NodeStatus{
				Addresses: []kube_core.NodeAddress{
					{Type: kube_core.NodeInternalIP, Address: "192.168.1.1"},
				},
			},
		}
		kubeClient := kube_client_fake.NewClientBuilder().
			WithScheme(k8sClientScheme).
			WithObjects(newNamespace(""), svc, newReadyEndpointSlice(), node).
			Build()

		Expect(reconcileMZA(kubeClient, recorder)).To(Succeed())

		mzas := listMZAs(kubeClient)
		Expect(mzas.Items).To(HaveLen(1))
		Expect(mzas.Items[0].Spec.Address).To(Equal("192.168.1.1"))
		Expect(mzas.Items[0].Spec.Port).To(Equal(int32(30001)))
	})

	It("creates MeshZoneAddress from externalIPs (takes precedence over service type)", func() {
		svc := newZoneProxyService(kube_core.ServiceTypeClusterIP, nil)
		svc.Spec.Ports = []kube_core.ServicePort{{Port: 10001}}
		svc.Spec.ExternalIPs = []string{"9.8.7.6", "0.0.0.1"}
		kubeClient := kube_client_fake.NewClientBuilder().
			WithScheme(k8sClientScheme).
			WithObjects(newNamespace(""), svc, newReadyEndpointSlice()).
			Build()

		Expect(reconcileMZA(kubeClient, recorder)).To(Succeed())

		mzas := listMZAs(kubeClient)
		Expect(mzas.Items).To(HaveLen(1))
		Expect(mzas.Items[0].Spec.Address).To(Equal("9.8.7.6"))
		Expect(mzas.Items[0].Spec.Port).To(Equal(int32(10001)))
	})

	It("emits warning and creates no MeshZoneAddress for ClusterIP without externalIPs", func() {
		svc := newZoneProxyService(kube_core.ServiceTypeClusterIP, nil)
		svc.Spec.Ports = []kube_core.ServicePort{{Port: 10001}}
		kubeClient := kube_client_fake.NewClientBuilder().
			WithScheme(k8sClientScheme).
			WithObjects(newNamespace(""), svc, newReadyEndpointSlice()).
			Build()

		Expect(reconcileMZA(kubeClient, recorder)).To(Succeed())
		Expect(listMZAs(kubeClient).Items).To(BeEmpty())
		Expect(recorder.Events).To(Receive(ContainSubstring("unable to determine public address")))
	})

	It("reads mesh name from namespace label when service has no mesh label", func() {
		svc := newZoneProxyService(kube_core.ServiceTypeLoadBalancer, nil) // no mesh label on service
		svc.Spec.Ports = []kube_core.ServicePort{{Port: 10001}}
		svc.Status = kube_core.ServiceStatus{
			LoadBalancer: kube_core.LoadBalancerStatus{
				Ingress: []kube_core.LoadBalancerIngress{{Hostname: "lb.example.com"}},
			},
		}
		kubeClient := kube_client_fake.NewClientBuilder().
			WithScheme(k8sClientScheme).
			WithObjects(newNamespace("prod"), svc, newReadyEndpointSlice()).
			Build()

		Expect(reconcileMZA(kubeClient, recorder)).To(Succeed())

		mzas := listMZAs(kubeClient)
		Expect(mzas.Items).To(HaveLen(1))
		Expect(mzas.Items[0].Labels[mesh_proto.MeshTag]).To(Equal("prod"))
	})

	It("falls back to default mesh when neither service nor namespace has mesh label", func() {
		svc := newZoneProxyService(kube_core.ServiceTypeLoadBalancer, nil)
		svc.Spec.Ports = []kube_core.ServicePort{{Port: 10001}}
		svc.Status = kube_core.ServiceStatus{
			LoadBalancer: kube_core.LoadBalancerStatus{
				Ingress: []kube_core.LoadBalancerIngress{{Hostname: "lb.example.com"}},
			},
		}
		kubeClient := kube_client_fake.NewClientBuilder().
			WithScheme(k8sClientScheme).
			WithObjects(newNamespace(""), svc, newReadyEndpointSlice()).
			Build()

		Expect(reconcileMZA(kubeClient, recorder)).To(Succeed())

		mzas := listMZAs(kubeClient)
		Expect(mzas.Items).To(HaveLen(1))
		Expect(mzas.Items[0].Labels[mesh_proto.MeshTag]).To(Equal("default"))
	})

	It("deletes existing MeshZoneAddress when service label is removed", func() {
		svc := &kube_core.Service{
			ObjectMeta: kube_meta.ObjectMeta{Name: testSvcName, Namespace: testNamespace},
			// no zone-proxy-type label
			Spec: kube_core.ServiceSpec{Type: kube_core.ServiceTypeLoadBalancer},
		}
		existingMZA := &meshzoneaddress_k8s.MeshZoneAddress{
			ObjectMeta: kube_meta.ObjectMeta{Name: testSvcName, Namespace: testNamespace},
		}
		kubeClient := kube_client_fake.NewClientBuilder().
			WithScheme(k8sClientScheme).
			WithObjects(newNamespace(""), svc, existingMZA).
			Build()

		Expect(reconcileMZA(kubeClient, recorder)).To(Succeed())
		Expect(listMZAs(kubeClient).Items).To(BeEmpty())
	})

	It("updates existing MeshZoneAddress when service changes", func() {
		svc := newZoneProxyService(kube_core.ServiceTypeLoadBalancer, nil)
		svc.Spec.Ports = []kube_core.ServicePort{{Port: 10001}}
		svc.Status = kube_core.ServiceStatus{
			LoadBalancer: kube_core.LoadBalancerStatus{
				Ingress: []kube_core.LoadBalancerIngress{{Hostname: "new-lb.example.com"}},
			},
		}
		existingMZA := &meshzoneaddress_k8s.MeshZoneAddress{
			ObjectMeta: kube_meta.ObjectMeta{Name: testSvcName, Namespace: testNamespace},
		}
		kubeClient := kube_client_fake.NewClientBuilder().
			WithScheme(k8sClientScheme).
			WithObjects(newNamespace(""), svc, newReadyEndpointSlice(), existingMZA).
			Build()

		Expect(reconcileMZA(kubeClient, recorder)).To(Succeed())

		mzas := listMZAs(kubeClient)
		Expect(mzas.Items).To(HaveLen(1))
		Expect(mzas.Items[0].Spec.Address).To(Equal("new-lb.example.com"))
	})
})
