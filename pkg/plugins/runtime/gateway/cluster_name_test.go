package gateway_test

import (
	envoy_cluster_v3 "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/pkg/plugins/runtime/gateway/route"
)

var _ = Describe("Cluster name", func() {
	NameOf := func(tags ...string) string {

		// Tags have to come in pairs.
		Expect(len(tags) % 2).To(BeZero())

		d := route.Destination{
			Destination: map[string]string{},
		}

		for i := 0; i < len(tags); i += 2 {
			d.Destination[tags[i]] = tags[i+1]
		}

		c := envoy_cluster_v3.Cluster{}
		name, err := route.DestinationClusterName(&d, &c)
		Expect(err).To(Succeed())

		return name
	}

	It("should require the service tag", func() {
		_, err := route.DestinationClusterName(
			&route.Destination{
				Destination: map[string]string{
					"somemthing":   "else",
					"kuma.io/zone": "one",
				},
			},
			&envoy_cluster_v3.Cluster{},
		)

		Expect(err).To(Not(Succeed()))
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

	It("should generate different names for different config", func() {
		n1, err := route.DestinationClusterName(
			&route.Destination{
				Destination: map[string]string{
					"kuma.io/service": "one",
				},
			},
			&envoy_cluster_v3.Cluster{RespectDnsTtl: true},
		)
		Expect(err).To(Succeed())

		n2, err := route.DestinationClusterName(
			&route.Destination{
				Destination: map[string]string{
					"kuma.io/service": "one",
				},
			},
			&envoy_cluster_v3.Cluster{},
		)
		Expect(err).To(Succeed())

		Expect(n1).ToNot(Equal(n2))
	})
})
