package server

import (
	konvoy_mesh "github.com/Kong/konvoy/components/konvoy-control-plane/model/api/v1alpha1"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/envoyproxy/go-control-plane/pkg/cache"

	k8s_core "k8s.io/api/core/v1"
	k8s_meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("Reconcile", func() {
	Describe("reconciler", func() {

		var hasher hasher
		var logger logger

		It("should generate a Snaphot per Envoy Node", func() {
			// setup
			store := cache.NewSnapshotCache(true, hasher, logger)
			r := &reconciler{&templateSnapshotGenerator{
				Template: &konvoy_mesh.ProxyTemplate{
					Spec: konvoy_mesh.ProxyTemplateSpec{
						Sources: []konvoy_mesh.ProxyTemplateSource{
							{
								Profile: &konvoy_mesh.ProxyTemplateProfileSource{
									Name: "transparent-inbound-proxy",
								},
							},
							{
								Profile: &konvoy_mesh.ProxyTemplateProfileSource{
									Name: "transparent-outbound-proxy",
								},
							},
						},
					},
				},
			}, &simpleSnapshotCacher{hasher, store}}

			// given
			pod := &k8s_core.Pod{
				ObjectMeta: k8s_meta.ObjectMeta{
					Name:      "app",
					Namespace: "example",
				},
				Spec: k8s_core.PodSpec{
					Containers: []k8s_core.Container{{
						Ports: []k8s_core.ContainerPort{{
							ContainerPort: 8080,
						}},
					}},
				},
				Status: k8s_core.PodStatus{
					PodIP: "192.168.0.1",
				},
			}

			// when
			err := r.OnUpdate(pod)
			Expect(err).ToNot(HaveOccurred())

			// then
			Eventually(func() bool {
				_, err := store.GetSnapshot("app.example")
				return err == nil
			}, "1s", "1ms").Should(BeTrue())
		})
	})
})
