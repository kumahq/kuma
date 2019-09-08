package util_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/Kong/kuma/pkg/plugins/discovery/k8s/util"

	kube_core "k8s.io/api/core/v1"
	kube_meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("Util", func() {
	Describe("MatchServiceThatSelectsPod", func() {
		It("should match", func() {
			// given
			pod := &kube_core.Pod{
				ObjectMeta: kube_meta.ObjectMeta{
					Labels: map[string]string{
						"app":               "demo-app",
						"pod-template-hash": "7cbbd658d5",
					},
				},
			}
			// and
			svc := &kube_core.Service{
				Spec: kube_core.ServiceSpec{
					Selector: map[string]string{
						"app": "demo-app",
					},
				},
			}

			// when
			predicate := MatchServiceThatSelectsPod(pod)
			// then
			Expect(predicate(svc)).To(BeTrue())
		})

		It("should not match", func() {
			// given
			pod := &kube_core.Pod{
				ObjectMeta: kube_meta.ObjectMeta{
					Labels: map[string]string{
						"app":               "demo-app",
						"pod-template-hash": "7cbbd658d5",
					},
				},
			}
			// and
			svc := &kube_core.Service{
				Spec: kube_core.ServiceSpec{
					Selector: map[string]string{
						"app": "nginx",
					},
				},
			}

			// when
			predicate := MatchServiceThatSelectsPod(pod)
			// then
			Expect(predicate(svc)).To(BeFalse())
		})
	})

	Describe("FindServices", func() {
		It("should match Services by a predicate", func() {
			// given
			pod := &kube_core.Pod{
				ObjectMeta: kube_meta.ObjectMeta{
					Labels: map[string]string{
						"app":               "demo-app",
						"pod-template-hash": "7cbbd658d5",
					},
				},
			}
			// and
			svcs := &kube_core.ServiceList{
				Items: []kube_core.Service{
					{
						ObjectMeta: kube_meta.ObjectMeta{
							Name: "demo-app",
						},
						Spec: kube_core.ServiceSpec{
							Selector: map[string]string{
								"app": "demo-app",
							},
						},
					},
					{
						ObjectMeta: kube_meta.ObjectMeta{
							Name: "nginx",
						},
						Spec: kube_core.ServiceSpec{
							Selector: map[string]string{
								"app": "nginx",
							},
						},
					},
				},
			}

			// when
			matchingServices := FindServices(svcs, MatchServiceThatSelectsPod(pod))
			// then
			Expect(matchingServices).To(HaveLen(1))
			Expect(matchingServices).To(ConsistOf(&svcs.Items[0]))
		})
	})

	Describe("CopyStringMap", func() {
		It("should return nil if input is nil", func() {
			Expect(CopyStringMap(nil)).To(BeNil())
		})
		It("should return empty map if input is empty map", func() {
			Expect(CopyStringMap(map[string]string{})).To(Equal(map[string]string{}))
		})
		It("should return a copy if input map is not empty", func() {
			// given
			original := map[string]string{
				"a": "b",
				"c": "d",
			}

			// when
			copy := CopyStringMap(original)
			// then
			Expect(copy).To(Equal(original))
			Expect(copy).ToNot(BeIdenticalTo(original))
		})
	})
})
