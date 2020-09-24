package metadata_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/pkg/plugins/runtime/k8s/metadata"
)

var _ = Describe("Kubernetes Annotations", func() {

	Context("GetEnabled()", func() {
		It("should parse value to bool", func() {
			annotations := map[string]string{
				"key1": "enabled",
				"key2": "disabled",
			}
			enabled, exist, err := metadata.Annotations(annotations).GetEnabled("key1")
			Expect(err).ToNot(HaveOccurred())
			Expect(enabled).To(BeTrue())
			Expect(exist).To(BeTrue())

			enabled, exist, err = metadata.Annotations(annotations).GetEnabled("key2")
			Expect(err).ToNot(HaveOccurred())
			Expect(enabled).To(BeFalse())
			Expect(exist).To(BeTrue())
		})

		It("should return error if value is wrong", func() {
			annotations := map[string]string{
				"key1": "not-enabled-at-all",
			}
			enabled, exist, err := metadata.Annotations(annotations).GetEnabled("key1")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("annotation \"key1\" has wrong value \"not-enabled-at-all\", available values are: \"enabled\", \"disabled\""))
			Expect(enabled).To(BeFalse())
			Expect(exist).To(BeTrue())
		})
	})
})
