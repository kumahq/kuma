package tags_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/pkg/xds/envoy/tags"
)

var _ = Describe("Cluster name", func() {
	NameOf := func(kv ...string) string {
		// Tags have to come in pairs.
		Expect(len(kv) % 2).To(BeZero())

		t := tags.Tags(map[string]string{})
		for i := 0; i < len(kv); i += 2 {
			t[kv[i]] = kv[i+1]
		}

		name, err := t.DestinationClusterName(nil)
		Expect(err).To(Succeed())

		return name
	}

	It("should require the service tag", func() {
		_, err := tags.Tags(map[string]string{
			"somemthing":   "else",
			"kuma.io/zone": "one",
		}).DestinationClusterName(nil)

		Expect(err).To(Not(Succeed()))
	})

	It("with only service tag don't do a hash", func() {
		res, err := tags.Tags(map[string]string{
			"kuma.io/service": "one",
		}).DestinationClusterName(nil)

		Expect(err).To(Succeed())
		Expect(res).To(Equal("one"))
	})

	It("should start with the service name", func() {
		Expect(NameOf(
			"kuma.io/service", "my-great-service",
			"kuma.io/zone", "one",
			"organization", "kumahq",
			"weekday", "friday",
			"weather", "sunny",
		)).To(HavePrefix("my-great-service"))
	})

	It("should generate the same name for the same tags", func() {
		Expect(NameOf(
			"kuma.io/service", "echo",
			"kuma.io/zone", "one",
		)).To(Equal(NameOf(
			"kuma.io/service", "echo",
			"kuma.io/zone", "one",
		)))
	})

	It("should generate different names for different tags", func() {
		Expect(NameOf(
			"kuma.io/service", "echo",
			"kuma.io/zone", "one",
		)).To(Not(Equal(NameOf(
			"kuma.io/service", "echo",
			"kuma.io/zone", "two",
		))))

		Expect(NameOf(
			"kuma.io/service", "echo",
			"kuma.io/zone", "one",
		)).To(Not(Equal(NameOf(
			"kuma.io/service", "echo",
			"kuma.io/zone", "one",
			"organization", "kumahq",
		))))
	})

	It("should generate different names for different tags", func() {
		n1, err := tags.Tags(map[string]string{
			"kuma.io/service": "one",
		}).DestinationClusterName(map[string]string{
			"tag1": "value1",
			"tag2": "value2",
		})
		Expect(err).To(Succeed())

		n2, err := tags.Tags(map[string]string{
			"kuma.io/service": "one",
		}).DestinationClusterName(map[string]string{
			"tag1": "value3",
			"abc":  "value2",
		})
		Expect(err).To(Succeed())

		Expect(n1).ToNot(Equal(n2))
	})
})
