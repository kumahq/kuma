package cni

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("shouldExcludePod", func() {
	It("should exclude pods that are present in excludedNamespaces", func() {
		// given
		excludeNamespaces := []string{"other-excluded-namespace", "my-excluded-namespace"}

		// when
		result := shouldExcludePod(excludeNamespaces, "my-excluded-namespace")

		// then
		Expect(result).To(BeTrue())
	})

	It("should not exclude pods that are not present in excludedNamespaces", func() {
		// given
		excludeNamespaces := []string{"other-excluded-namespace", "my-excluded-namespace"}

		// when
		result := shouldExcludePod(excludeNamespaces, "not-excluded-namespace")

		// then
		Expect(result).To(BeFalse())
	})
})

var _ = Describe("isInitContainerPresent", func() {
	It("should exclude pods that have init container", func() {
		// given
		initContainersMap := map[string]struct{}{"kuma-init": {}, "init-container": {}}

		// when
		result := isInitContainerPresent(initContainersMap)

		// then
		Expect(result).To(BeTrue())
	})

	It("should not exclude pods that do not have init container", func() {
		// given
		initContainersMap := map[string]struct{}{"some-other-init-container": {}}

		// when
		result := isInitContainerPresent(initContainersMap)

		// then
		Expect(result).To(BeFalse())
	})
})

var _ = Describe("excludeByMissingSidecarInjectedAnnotation", func() {
	It("should not exclude pods that have sidecar injected annotation", func() {
		// given
		annotations := map[string]string{"kuma.io/sidecar-injected": "true"}

		// when
		result := excludeByMissingSidecarInjectedAnnotation(annotations)

		// then
		Expect(result).To(BeFalse())
	})

	It("should exclude pods that do have sidecar injected annotation set to false", func() {
		// given
		annotations := map[string]string{"kuma.io/sidecar-injected": "false"}

		// when
		result := excludeByMissingSidecarInjectedAnnotation(annotations)

		// then
		Expect(result).To(BeTrue())
	})

	It("should exclude pods that do not have sidecar injected annotation set", func() {
		// given
		annotations := map[string]string{"kuma.io/some-annotation": "true"}

		// when
		result := excludeByMissingSidecarInjectedAnnotation(annotations)

		// then
		Expect(result).To(BeTrue())
	})
})
