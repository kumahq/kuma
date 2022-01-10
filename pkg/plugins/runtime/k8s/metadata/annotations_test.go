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

	Context("GetUint32()", func() {
		It("should parse value to uint32", func() {
			// given
			annotations := map[string]string{
				"key1": "100",
			}

			val, hasKey, err := metadata.Annotations(annotations).GetUint32("key1")
			Expect(err).ToNot(HaveOccurred())
			Expect(hasKey).To(Equal(true))
			Expect(val).To(Equal(uint32(100)))
		})

		It("should return error if value has wrong format", func() {
			annotations := map[string]string{
				"key1": "dummy",
			}

			_, _, err := metadata.Annotations(annotations).GetUint32("key1")
			Expect(err.Error()).To(ContainSubstring("failed to parse annotation \"key1\": strconv.ParseUint: parsing \"dummy\": invalid syntax"))
		})
	})

	Context("GetBool()", func() {
		It("should parse value to bool", func() {
			annotations := map[string]string{
				"key1": "true",
			}

			val, hasKey, err := metadata.Annotations(annotations).GetBool("key1")
			Expect(err).ToNot(HaveOccurred())
			Expect(hasKey).To(Equal(true))
			Expect(val).To(Equal(true))
		})

		It("should return error if value has wrong format", func() {
			annotations := map[string]string{
				"key1": "dummy",
			}

			_, _, err := metadata.Annotations(annotations).GetBool("key1")
			Expect(err.Error()).To(ContainSubstring("failed to parse annotation \"key1\": strconv.ParseBool: parsing \"dummy\": invalid syntax"))
		})
	})

	Context("GetMap()", func() {
		It("should parse value to map", func() {
			// given
			annotations := map[string]string{
				"key1": "TEST1=1;TEST2=2",
			}

			// when
			m, err := metadata.Annotations(annotations).GetMap("key1")

			// then
			Expect(err).ToNot(HaveOccurred())
			Expect(m).To(Equal(map[string]string{"TEST1": "1", "TEST2": "2"}))
		})

		It("should return error if value has wrong format", func() {
			// given
			annotations := map[string]string{
				"key1": "TESTTEST",
			}

			// when
			_, err := metadata.Annotations(annotations).GetMap("key1")

			// then
			Expect(err).To(MatchError(`invalid format. Map in "key1" has to be provided in the following format: key1=value1;key2=value2`))
		})
	})
})
