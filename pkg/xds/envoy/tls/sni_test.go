package tls_test

import (
	"fmt"

	"github.com/asaskevich/govalidator"
	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	meshmzservice_api "github.com/kumahq/kuma/v3/pkg/core/resources/apis/meshmultizoneservice/api/v1alpha1"
	meshservice_api "github.com/kumahq/kuma/v3/pkg/core/resources/apis/meshservice/api/v1alpha1"
	"github.com/kumahq/kuma/v3/pkg/core/resources/model"
	"github.com/kumahq/kuma/v3/pkg/xds/envoy/tls"
)

var _ = Describe("SNI", func() {
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
		for i := range 100_000 {
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
		for range 100_000 {
			sni := tls.SNIForResource(serviceName+uuid.New().String(), "demo", meshservice_api.MeshServiceType, 8080, nil)
			_, ok := snis[sni]
			Expect(ok).To(BeFalse())
			snis[sni] = struct{}{}
		}
	})
})
