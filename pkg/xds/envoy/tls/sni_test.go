package tls_test

import (
	"fmt"

	"github.com/asaskevich/govalidator"
	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	meshmzservice_api "github.com/kumahq/kuma/pkg/core/resources/apis/meshmultizoneservice/api/v1alpha1"
	meshservice_api "github.com/kumahq/kuma/pkg/core/resources/apis/meshservice/api/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	envoy_tags "github.com/kumahq/kuma/pkg/xds/envoy/tags"
	"github.com/kumahq/kuma/pkg/xds/envoy/tls"
)

var _ = Describe("SNI", func() {
	It("should convert tags to SNI", func() {
		// given
		tags := map[string]string{
			"kuma.io/service": "backend",
			"version":         "v1",
			"env":             "prod",
			"region":          "eu",
			"app":             "backend-app",
		}
		expected := "backend{app=backend-app,env=prod,region=eu,version=v1}"

		// when
		actual := tls.SNIFromTags(tags)

		// then
		Expect(actual).To(Equal(expected))
	})

	It("should convert tags to SNI with only service name", func() {
		// given
		tags := map[string]string{
			"kuma.io/service": "backend",
		}
		expected := "backend"

		// when
		actual := tls.SNIFromTags(tags)

		// then
		Expect(actual).To(Equal(expected))
	})

	It("should convert SNI to tags", func() {
		// given
		sni := "backend{app=backend-app,env=prod,region=eu,version=v1}"
		expectedTags := envoy_tags.Tags{
			"kuma.io/service": "backend",
			"version":         "v1",
			"env":             "prod",
			"region":          "eu",
			"app":             "backend-app",
		}

		// when
		tags, err := tls.TagsFromSNI(sni)

		// then
		Expect(err).ToNot(HaveOccurred())
		Expect(tags).To(Equal(expectedTags))
	})

	It("should convert SNI to tags with only service name", func() {
		// given
		sni := "backend"
		expectedTags := envoy_tags.Tags{
			"kuma.io/service": "backend",
		}

		// when
		tags, err := tls.TagsFromSNI(sni)

		// then
		Expect(err).ToNot(HaveOccurred())
		Expect(tags).To(Equal(expectedTags))
	})

	DescribeTable("should fail when converting SNI to tags", func(sni string, errorMessage string) {
		// when
		_, err := tls.TagsFromSNI(sni)

		// then
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(Equal(errorMessage))
	},
		Entry("broken tags", "backend{", "invalid format of tags, pairs should be separated by , and key should be separated from value by ="),
		Entry("to many separators", "backend{mesh=default{mesh", "cannot parse tags from sni: backend{mesh=default{mesh"),
	)

	type testCase struct {
		resName        string
		meshName       string
		resType        model.ResourceType
		port           int32
		additionalData map[string]string
		expected       string
	}
	DescribeTable("should convert SNI for resource",
		func(given testCase) {
			sni := tls.SNIForResource(
				given.resName,
				given.meshName,
				given.resType,
				given.port,
				given.additionalData,
			)

			Expect(sni).To(Equal(given.expected))
			Expect(govalidator.IsDNSName(sni)).To(BeTrue())
		},
		Entry("simple", testCase{
			resName:        "backend",
			meshName:       "demo",
			resType:        meshservice_api.MeshServiceType,
			port:           8080,
			additionalData: nil,
			expected:       "ae10a8071b8a8eeb8.backend.8080.demo.ms",
		}),
		Entry("simple subset", testCase{
			resName:  "backend",
			meshName: "demo",
			resType:  meshservice_api.MeshServiceType,
			port:     8080,
			additionalData: map[string]string{
				"x": "a",
			},
			expected: "a333a125865a97632.backend.8080.demo.ms",
		}),
		Entry("going over limit", testCase{
			resName:  "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa-qwe",
			meshName: "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb-qwe",
			resType:  meshservice_api.MeshServiceType,
			port:     8080,
			additionalData: map[string]string{
				"x": "a",
			},
			expected: "a5b91d8a08567bf09.aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaax.8080.bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbx.ms",
		}),
		Entry("mesh multizone service", testCase{
			resName:        "backend",
			meshName:       "demo",
			resType:        meshmzservice_api.MeshMultiZoneServiceType,
			port:           8080,
			additionalData: nil,
			expected:       "ae10a8071b8a8eeb8.backend.8080.demo.mzms",
		}),
	)

	It("SNI hash does not easily collide of the same services with different tags", func() {
		snis := map[string]struct{}{}
		for i := 0; i < 100_000; i++ {
			sni := tls.SNIForResource("backend", "demo", meshservice_api.MeshServiceType, 8080, map[string]string{
				"version": fmt.Sprintf("%d", i),
			})
			_, ok := snis[sni]
			Expect(ok).To(BeFalse())
			snis[sni] = struct{}{}
		}
	})

	It("SNI hash does not easily collide of the services with very long names", func() {
		snis := map[string]struct{}{}
		serviceName := "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
		for i := 0; i < 100_000; i++ {
			sni := tls.SNIForResource(serviceName+uuid.New().String(), "demo", meshservice_api.MeshServiceType, 8080, nil)
			_, ok := snis[sni]
			Expect(ok).To(BeFalse())
			snis[sni] = struct{}{}
		}
	})
})
