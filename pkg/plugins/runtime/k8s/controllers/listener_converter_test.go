package controllers_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	kube_core "k8s.io/api/core/v1"
	kube_meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	kube_intstr "k8s.io/apimachinery/pkg/util/intstr"

	mesh_proto "github.com/kumahq/kuma/v2/api/mesh/v1alpha1"
	. "github.com/kumahq/kuma/v2/pkg/plugins/runtime/k8s/controllers"
	"github.com/kumahq/kuma/v2/pkg/plugins/runtime/k8s/metadata"
)

var _ = Describe("ListenersForService", func() {
	pod := func(podIP string, containerStatuses []kube_core.ContainerStatus) *kube_core.Pod {
		return &kube_core.Pod{
			ObjectMeta: kube_meta.ObjectMeta{
				Namespace: "demo",
				Name:      "test-pod",
			},
			Spec: kube_core.PodSpec{
				Containers: []kube_core.Container{
					{
						Name: "app",
						Ports: []kube_core.ContainerPort{
							{ContainerPort: 10001},
						},
					},
				},
			},
			Status: kube_core.PodStatus{
				PodIP:             podIP,
				ContainerStatuses: containerStatuses,
			},
		}
	}

	svc := func(name, proxyType string, port int32) *kube_core.Service {
		labels := map[string]string{}
		if proxyType != "" {
			labels[metadata.KumaZoneProxyTypeLabel] = proxyType
		}
		return &kube_core.Service{
			ObjectMeta: kube_meta.ObjectMeta{
				Namespace: "demo",
				Name:      name,
				Labels:    labels,
			},
			Spec: kube_core.ServiceSpec{
				Ports: []kube_core.ServicePort{
					{
						Name:       "main",
						Protocol:   kube_core.ProtocolTCP,
						Port:       port,
						TargetPort: kube_intstr.FromInt32(port),
					},
				},
			},
		}
	}

	readyStatuses := []kube_core.ContainerStatus{
		{Name: "app", Ready: true},
	}

	notReadySidecar := []kube_core.ContainerStatus{
		{Name: "app", Ready: true},
		{Name: "kuma-sidecar", Ready: false},
	}

	DescribeTable("should return nil for services without the zone-proxy-type label",
		func() {
			listeners, err := ListenersForService(pod("192.168.0.1", readyStatuses), svc("backend", "", 8080))
			Expect(err).ToNot(HaveOccurred())
			Expect(listeners).To(BeNil())
		},
		Entry("no label"),
	)

	DescribeTable("should return an error for an invalid label value",
		func() {
			_, err := ListenersForService(pod("192.168.0.1", readyStatuses), svc("bad", "invalid", 10001))
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("invalid"))
		},
		Entry("unknown zone-proxy-type"),
	)

	DescribeTable("should generate a ZoneIngress listener",
		func() {
			listeners, err := ListenersForService(pod("192.168.0.1", readyStatuses), svc("zi", KumaZoneProxyTypeIngress, 10001))
			Expect(err).ToNot(HaveOccurred())
			Expect(listeners).To(HaveLen(1))
			l := listeners[0]
			Expect(l.Type).To(Equal(mesh_proto.Dataplane_Networking_Listener_ZoneIngress))
			Expect(l.Address).To(Equal("192.168.0.1"))
			Expect(l.Port).To(Equal(uint32(10001)))
			Expect(l.Name).To(Equal("main"))
			Expect(l.State).To(Equal(mesh_proto.Dataplane_Networking_Listener_Ready))
		},
		Entry("ZoneIngress ready"),
	)

	DescribeTable("should generate a ZoneEgress listener",
		func() {
			listeners, err := ListenersForService(pod("10.0.0.5", readyStatuses), svc("ze", KumaZoneProxyTypeEgress, 10002))
			Expect(err).ToNot(HaveOccurred())
			Expect(listeners).To(HaveLen(1))
			l := listeners[0]
			Expect(l.Type).To(Equal(mesh_proto.Dataplane_Networking_Listener_ZoneEgress))
			Expect(l.Address).To(Equal("10.0.0.5"))
			Expect(l.Port).To(Equal(uint32(10002)))
			Expect(l.State).To(Equal(mesh_proto.Dataplane_Networking_Listener_Ready))
		},
		Entry("ZoneEgress ready"),
	)

	DescribeTable("should set NotReady when kuma-sidecar is not ready",
		func() {
			listeners, err := ListenersForService(pod("192.168.0.1", notReadySidecar), svc("zi", KumaZoneProxyTypeIngress, 10001))
			Expect(err).ToNot(HaveOccurred())
			Expect(listeners).To(HaveLen(1))
			Expect(listeners[0].State).To(Equal(mesh_proto.Dataplane_Networking_Listener_NotReady))
		},
		Entry("sidecar not ready"),
	)

	DescribeTable("should set NotReady when pod is terminating",
		func() {
			p := pod("192.168.0.1", readyStatuses)
			now := kube_meta.Now()
			p.DeletionTimestamp = &now
			listeners, err := ListenersForService(p, svc("zi", KumaZoneProxyTypeIngress, 10001))
			Expect(err).ToNot(HaveOccurred())
			Expect(listeners).To(HaveLen(1))
			Expect(listeners[0].State).To(Equal(mesh_proto.Dataplane_Networking_Listener_NotReady))
		},
		Entry("pod terminating"),
	)

	DescribeTable("should skip non-TCP ports",
		func() {
			s := svc("zi", KumaZoneProxyTypeIngress, 10001)
			s.Spec.Ports[0].Protocol = kube_core.ProtocolUDP
			listeners, err := ListenersForService(pod("192.168.0.1", readyStatuses), s)
			Expect(err).ToNot(HaveOccurred())
			Expect(listeners).To(BeEmpty())
		},
		Entry("UDP port skipped"),
	)

	DescribeTable("should use svcName-port fallback name when port has no name",
		func() {
			s := svc("zone-ingress", KumaZoneProxyTypeIngress, 10001)
			s.Spec.Ports[0].Name = ""
			listeners, err := ListenersForService(pod("192.168.0.1", readyStatuses), s)
			Expect(err).ToNot(HaveOccurred())
			Expect(listeners).To(HaveLen(1))
			Expect(listeners[0].Name).To(Equal("zone-ingress-10001"))
		},
		Entry("fallback name"),
	)
})
