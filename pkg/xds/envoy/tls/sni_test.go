package tls_test

import (
	"fmt"
	"strings"

	"github.com/asaskevich/govalidator"
	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/v2/pkg/core/kri"
	meshexternalservice_api "github.com/kumahq/kuma/v2/pkg/core/resources/apis/meshexternalservice/api/v1alpha1"
	meshmzservice_api "github.com/kumahq/kuma/v2/pkg/core/resources/apis/meshmultizoneservice/api/v1alpha1"
	meshservice_api "github.com/kumahq/kuma/v2/pkg/core/resources/apis/meshservice/api/v1alpha1"
	"github.com/kumahq/kuma/v2/pkg/core/resources/model"
	envoy_tags "github.com/kumahq/kuma/v2/pkg/xds/envoy/tags"
	"github.com/kumahq/kuma/v2/pkg/xds/envoy/tls"
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

	type kriTestCase struct {
		id        kri.Identifier
		expected  string
		expectErr bool
	}
	DescribeTable("SNIFromKRI",
		func(tc kriTestCase) {
			sni, err := tls.SNIFromKRI(tc.id)
			if tc.expectErr {
				Expect(err).To(HaveOccurred())
			} else {
				Expect(err).ToNot(HaveOccurred())
				Expect(sni).To(Equal(tc.expected))
				Expect(govalidator.IsDNSName(sni)).To(BeTrue())
			}
		},
		// MeshService — global (no zone, no namespace)
		Entry("MeshService global", kriTestCase{
			id: kri.Identifier{
				ResourceType: meshservice_api.MeshServiceType,
				Mesh:         "default",
				Name:         "backend",
				SectionName:  "http",
			},
			expected: "sni.msvc.default.backend.http",
		}),
		// MeshService — zone-originated, system namespace (zone set, no namespace)
		Entry("MeshService zone system-ns", kriTestCase{
			id: kri.Identifier{
				ResourceType: meshservice_api.MeshServiceType,
				Mesh:         "default",
				Zone:         "east",
				Name:         "backend",
				SectionName:  "http",
			},
			expected: "sni.msvc.default.east.backend.http",
		}),
		// MeshService — zone-originated, custom namespace
		Entry("MeshService zone custom-ns", kriTestCase{
			id: kri.Identifier{
				ResourceType: meshservice_api.MeshServiceType,
				Mesh:         "default",
				Zone:         "east",
				Namespace:    "app-ns",
				Name:         "backend",
				SectionName:  "http",
			},
			expected: "sni.msvc.default.east.app-ns.backend.http",
		}),
		// MeshExternalService — global
		Entry("MeshExternalService global", kriTestCase{
			id: kri.Identifier{
				ResourceType: meshexternalservice_api.MeshExternalServiceType,
				Mesh:         "default",
				Name:         "ext-backend",
				SectionName:  "9000",
			},
			expected: "sni.extsvc.default.ext-backend.9000",
		}),
		// MeshExternalService — zone-originated
		Entry("MeshExternalService zone", kriTestCase{
			id: kri.Identifier{
				ResourceType: meshexternalservice_api.MeshExternalServiceType,
				Mesh:         "prod",
				Zone:         "west",
				Name:         "ext-backend",
				SectionName:  "9000",
			},
			expected: "sni.extsvc.prod.west.ext-backend.9000",
		}),
		// MeshMultiZoneService — global (MeshMultiZoneService never carries zone/namespace)
		Entry("MeshMultiZoneService global", kriTestCase{
			id: kri.Identifier{
				ResourceType: meshmzservice_api.MeshMultiZoneServiceType,
				Mesh:         "default",
				Name:         "global-svc",
				SectionName:  "http",
			},
			expected: "sni.mzsvc.default.global-svc.http",
		}),
		// MeshMultiZoneService — numeric port as sectionName
		Entry("MeshMultiZoneService numeric port", kriTestCase{
			id: kri.Identifier{
				ResourceType: meshmzservice_api.MeshMultiZoneServiceType,
				Mesh:         "default",
				Name:         "global-svc",
				SectionName:  "8080",
			},
			expected: "sni.mzsvc.default.global-svc.8080",
		}),
		// Error: empty Mesh
		Entry("error empty mesh", kriTestCase{
			id: kri.Identifier{
				ResourceType: meshservice_api.MeshServiceType,
				Name:         "backend",
				SectionName:  "http",
			},
			expectErr: true,
		}),
		// Error: empty Name
		Entry("error empty name", kriTestCase{
			id: kri.Identifier{
				ResourceType: meshservice_api.MeshServiceType,
				Mesh:         "default",
				SectionName:  "http",
			},
			expectErr: true,
		}),
		// Error: empty SectionName
		Entry("error empty sectionName", kriTestCase{
			id: kri.Identifier{
				ResourceType: meshservice_api.MeshServiceType,
				Mesh:         "default",
				Name:         "backend",
			},
			expectErr: true,
		}),
		// Error: namespace set without zone
		Entry("error namespace without zone", kriTestCase{
			id: kri.Identifier{
				ResourceType: meshservice_api.MeshServiceType,
				Mesh:         "default",
				Namespace:    "app-ns",
				Name:         "backend",
				SectionName:  "http",
			},
			expectErr: true,
		}),
		// Error: mesh segment contains '.'
		Entry("error mesh contains dot", kriTestCase{
			id: kri.Identifier{
				ResourceType: meshservice_api.MeshServiceType,
				Mesh:         "de.fault",
				Name:         "backend",
				SectionName:  "http",
			},
			expectErr: true,
		}),
		// Error: name segment contains '.'
		Entry("error name contains dot", kriTestCase{
			id: kri.Identifier{
				ResourceType: meshservice_api.MeshServiceType,
				Mesh:         "default",
				Name:         "back.end",
				SectionName:  "http",
			},
			expectErr: true,
		}),
		// Error: zone segment contains '.'
		Entry("error zone contains dot", kriTestCase{
			id: kri.Identifier{
				ResourceType: meshservice_api.MeshServiceType,
				Mesh:         "default",
				Zone:         "east.zone",
				Name:         "backend",
				SectionName:  "http",
			},
			expectErr: true,
		}),
		// Error: sectionName contains '.'
		Entry("error sectionName contains dot", kriTestCase{
			id: kri.Identifier{
				ResourceType: meshservice_api.MeshServiceType,
				Mesh:         "default",
				Name:         "backend",
				SectionName:  "http.port",
			},
			expectErr: true,
		}),
		// Length limits: 63-char label is allowed.
		Entry("label exactly 63 chars", kriTestCase{
			id: kri.Identifier{
				ResourceType: meshservice_api.MeshServiceType,
				Mesh:         "default",
				Name:         strings.Repeat("a", 63),
				SectionName:  "http",
			},
			expected: "sni.msvc.default." + strings.Repeat("a", 63) + ".http",
		}),
		// Error: 64-char label exceeds DNS label limit.
		Entry("error label exceeds 63 chars", kriTestCase{
			id: kri.Identifier{
				ResourceType: meshservice_api.MeshServiceType,
				Mesh:         "default",
				Name:         strings.Repeat("a", 64),
				SectionName:  "http",
			},
			expectErr: true,
		}),
		// Error: total > 253 across multiple max-length labels (≥4×64-1).
		Entry("error total exceeds 253 chars", kriTestCase{
			id: kri.Identifier{
				ResourceType: meshservice_api.MeshServiceType,
				Mesh:         strings.Repeat("a", 63),
				Zone:         strings.Repeat("b", 63),
				Namespace:    strings.Repeat("c", 63),
				Name:         strings.Repeat("d", 63),
				SectionName:  "http",
			},
			expectErr: true,
		}),
	)

	Describe("ValidateSNIForKRI", func() {
		It("returns nil for a valid identifier", func() {
			err := tls.ValidateSNIForKRI(kri.Identifier{
				ResourceType: meshservice_api.MeshServiceType,
				Mesh:         "default",
				Name:         "backend",
				SectionName:  "http",
			})
			Expect(err).ToNot(HaveOccurred())
		})

		It("returns an error when a segment exceeds the DNS label limit", func() {
			err := tls.ValidateSNIForKRI(kri.Identifier{
				ResourceType: meshservice_api.MeshServiceType,
				Mesh:         "default",
				Name:         strings.Repeat("a", 64),
				SectionName:  "http",
			})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("DNS label limit"))
		})

		It("returns an error when the total hostname exceeds the DNS hostname limit", func() {
			err := tls.ValidateSNIForKRI(kri.Identifier{
				ResourceType: meshservice_api.MeshServiceType,
				Mesh:         strings.Repeat("a", 63),
				Zone:         strings.Repeat("b", 63),
				Namespace:    strings.Repeat("c", 63),
				Name:         strings.Repeat("d", 63),
				SectionName:  "http",
			})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("DNS hostname limit"))
		})
	})
})
