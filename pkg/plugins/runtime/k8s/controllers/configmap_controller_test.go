package controllers_test

import (
	"context"
	"fmt"
	"sort"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	kube_core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kube_client "sigs.k8s.io/controller-runtime/pkg/client"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/log"
	"github.com/kumahq/kuma/pkg/plugins/resources/k8s"
	mesh_k8s "github.com/kumahq/kuma/pkg/plugins/resources/k8s/native/api/v1alpha1"
	"github.com/kumahq/kuma/pkg/plugins/runtime/k8s/controllers"
	"github.com/kumahq/kuma/pkg/plugins/runtime/k8s/metadata"
)

var _ = Describe("DataplaneToMeshMapper", func() {
	It("should map ingress to list of meshes", func() {
		l := log.NewLogger(log.InfoLevel)
		mapper := controllers.DataplaneToMeshMapper(l, "ns", k8s.NewSimpleConverter())
		requests := mapper(context.Background(), &mesh_k8s.Dataplane{
			Mesh: "mesh-1",
			Spec: mesh_k8s.ToSpec(&mesh_proto.Dataplane{
				Networking: &mesh_proto.Dataplane_Networking{
					Address: "10.20.1.2",
					Inbound: []*mesh_proto.Dataplane_Networking_Inbound{
						{
							Port: 10001,
							Tags: map[string]string{mesh_proto.ServiceTag: "redis"},
						},
					},
				},
			}),
		})
		requestsStr := []string{}
		for _, r := range requests {
			requestsStr = append(requestsStr, r.Name)
		}
		Expect(requestsStr).To(HaveLen(1))
		Expect(requestsStr).To(ConsistOf("kuma-mesh-1-dns-vips"))
	})
})

var _ = Describe("ServiceToConfigMapMapper", func() {
	ctx := context.Background()
	var nsName string
	defaultNs := kube_core.Namespace{
		ObjectMeta: metav1.ObjectMeta{},
	}
	AfterEach(func() {
		ns := kube_core.Namespace{}
		key := kube_client.ObjectKey{Name: nsName}
		Expect(k8sClient.Get(ctx, key, &ns)).To(Succeed())
		Expect(k8sClient.Delete(ctx, &ns)).To(Succeed())
	})
	BeforeEach(func() {
		nsName = fmt.Sprintf("srv-to-configmap-mapper-%d", time.Now().UnixMilli())
	})
	serviceFn := func(selector map[string]string) kube_core.Service {
		return kube_core.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name: "svc",
			},
			Spec: kube_core.ServiceSpec{
				Selector: selector,
				Ports: []kube_core.ServicePort{
					{Port: 80},
				},
			},
		}
	}
	podFn := func(name string, labels map[string]string, annotations map[string]string) kube_core.Pod {
		return kube_core.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:        name,
				Labels:      labels,
				Annotations: annotations,
			},
			Spec: kube_core.PodSpec{
				Containers: []kube_core.Container{
					{Name: "foo", Image: "busybox"},
				},
			},
		}
	}
	DescribeTable("should map services to list of config maps",
		func(givenService kube_core.Service, givenNamespace kube_core.Namespace, givenPods []kube_core.Pod, expectedMeshes []string) {
			l := log.NewLogger(log.InfoLevel)
			givenNamespace.Name = nsName
			Expect(k8sClient.Create(ctx, &givenNamespace)).To(Succeed())
			givenService.Namespace = givenNamespace.Name
			Expect(k8sClient.Create(ctx, &givenService)).To(Succeed())
			for i := range givenPods {
				pod := givenPods[i]
				pod.Namespace = givenNamespace.Name
				Expect(k8sClient.Create(ctx, &pod)).To(Succeed())
			}
			mapper := controllers.ServiceToConfigMapsMapper(k8sClient, l, "ns")
			requests := mapper(context.Background(), &givenService)
			requestsStr := []string{}
			for _, r := range requests {
				requestsStr = append(requestsStr, r.Name)
			}
			sort.Strings(requestsStr)
			Expect(requestsStr).To(Equal(expectedMeshes))
		},
		Entry("no pod match",
			serviceFn(map[string]string{"app": "app1"}),
			defaultNs,
			[]kube_core.Pod{
				podFn("pod1", map[string]string{"app": "app2"}, nil),
				podFn("pod2", map[string]string{"app": "app2", metadata.KumaMeshLabel: "mesh2"}, map[string]string{}),
			},
			[]string{},
		),
		Entry("namespace not annotated selects only matching pods",
			serviceFn(map[string]string{"app": "app1"}),
			defaultNs,
			[]kube_core.Pod{
				podFn("pod1", map[string]string{"app": "app1"}, nil),
				podFn("pod2", map[string]string{"app": "app2", metadata.KumaMeshLabel: "mesh2"}, map[string]string{}),
			},
			[]string{"kuma-default-dns-vips"},
		),
		Entry("namespace not annotated pod annotated selects only matching pods",
			serviceFn(map[string]string{"app": "app1"}),
			defaultNs,
			[]kube_core.Pod{
				podFn("pod1", map[string]string{"app": "app1", metadata.KumaMeshLabel: "mesh1"}, map[string]string{}),
				podFn("pod2", map[string]string{"app": "app2", metadata.KumaMeshLabel: "mesh2"}, map[string]string{}),
			},
			[]string{"kuma-mesh1-dns-vips"},
		),
		Entry("namespace not annotated pod annotated matches on multiple meshes",
			serviceFn(map[string]string{"app": "app1"}),
			defaultNs,
			[]kube_core.Pod{
				podFn("pod1", map[string]string{"app": "app1", metadata.KumaMeshLabel: "mesh1"}, map[string]string{}),
				podFn("pod2", map[string]string{"app": "app1", metadata.KumaMeshLabel: "mesh2"}, map[string]string{}),
			},
			[]string{"kuma-mesh1-dns-vips", "kuma-mesh2-dns-vips"},
		),
		Entry("namespace annotated pod not annotated",
			serviceFn(map[string]string{"app": "app1"}),
			kube_core.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						metadata.KumaMeshLabel: "mesh1",
					},
				},
			},
			[]kube_core.Pod{
				podFn("pod1", map[string]string{"app": "app1"}, nil),
				podFn("pod2", map[string]string{"app": "app1"}, nil),
			},
			[]string{"kuma-mesh1-dns-vips"},
		),
		Entry("namespace annotated pod annotated",
			serviceFn(map[string]string{"app": "app1"}),
			kube_core.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						metadata.KumaMeshLabel: "mesh1",
					},
				},
			},
			[]kube_core.Pod{
				podFn("pod1", map[string]string{"app": "app1"}, map[string]string{metadata.KumaMeshLabel: "mesh2"}),
				podFn("pod2", map[string]string{"app": "app1"}, nil),
			},
			[]string{"kuma-mesh1-dns-vips", "kuma-mesh2-dns-vips"},
		),
		Entry("namespace label pod has label",
			serviceFn(map[string]string{"app": "app1"}),
			kube_core.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						metadata.KumaMeshLabel: "mesh1",
					},
				},
			},
			[]kube_core.Pod{
				podFn("pod1", map[string]string{"app": "app1", metadata.KumaMeshLabel: "mesh1"}, map[string]string{}),
				podFn("pod2", map[string]string{"app": "app1"}, nil),
			},
			[]string{"kuma-mesh1-dns-vips"},
		),
		Entry("namespace label pod has annotation",
			serviceFn(map[string]string{"app": "app1"}),
			kube_core.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						metadata.KumaMeshLabel: "mesh1",
					},
				},
			},
			[]kube_core.Pod{
				podFn("pod1", map[string]string{"app": "app1"}, map[string]string{metadata.KumaMeshAnnotation: "mesh1"}),
				podFn("pod2", map[string]string{"app": "app1"}, nil),
			},
			[]string{"kuma-mesh1-dns-vips"},
		),
	)
})
