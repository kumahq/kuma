package sni_test

import (
	"strings"

	"github.com/asaskevich/govalidator"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/v2/pkg/core/kri"
	core_mesh "github.com/kumahq/kuma/v2/pkg/core/resources/apis/mesh"
	meshexternalservice_api "github.com/kumahq/kuma/v2/pkg/core/resources/apis/meshexternalservice/api/v1alpha1"
	meshmzservice_api "github.com/kumahq/kuma/v2/pkg/core/resources/apis/meshmultizoneservice/api/v1alpha1"
	meshservice_api "github.com/kumahq/kuma/v2/pkg/core/resources/apis/meshservice/api/v1alpha1"
	"github.com/kumahq/kuma/v2/pkg/core/resources/model"
	"github.com/kumahq/kuma/v2/pkg/core/resources/sni"
)

var _ = Describe("FromKRI / ValidateKRI", func() {
	type kriTestCase struct {
		id        kri.Identifier
		expected  string
		expectErr bool
	}
	DescribeTable("",
		func(tc kriTestCase) {
			errs := sni.ValidateKRI(tc.id)
			if tc.expectErr {
				Expect(errs).ToNot(BeEmpty())
				return
			}
			Expect(errs).To(BeEmpty())
			out := sni.FromKRI(tc.id)
			Expect(out).To(Equal(tc.expected))
			Expect(govalidator.IsDNSName(out)).To(BeTrue())
		},
		Entry("MeshService global", kriTestCase{
			id: kri.Identifier{
				ResourceType: meshservice_api.MeshServiceType,
				Mesh:         "default",
				Name:         "backend",
				SectionName:  "http",
			},
			expected: "sni.msvc.default.backend.http",
		}),
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
		Entry("MeshExternalService global", kriTestCase{
			id: kri.Identifier{
				ResourceType: meshexternalservice_api.MeshExternalServiceType,
				Mesh:         "default",
				Name:         "ext-backend",
				SectionName:  "9000",
			},
			expected: "sni.extsvc.default.ext-backend.9000",
		}),
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
		Entry("MeshMultiZoneService global", kriTestCase{
			id: kri.Identifier{
				ResourceType: meshmzservice_api.MeshMultiZoneServiceType,
				Mesh:         "default",
				Name:         "global-svc",
				SectionName:  "http",
			},
			expected: "sni.mzsvc.default.global-svc.http",
		}),
		Entry("MeshMultiZoneService numeric port", kriTestCase{
			id: kri.Identifier{
				ResourceType: meshmzservice_api.MeshMultiZoneServiceType,
				Mesh:         "default",
				Name:         "global-svc",
				SectionName:  "8080",
			},
			expected: "sni.mzsvc.default.global-svc.8080",
		}),
		Entry("error empty mesh", kriTestCase{
			id: kri.Identifier{
				ResourceType: meshservice_api.MeshServiceType,
				Name:         "backend",
				SectionName:  "http",
			},
			expectErr: true,
		}),
		Entry("error empty name", kriTestCase{
			id: kri.Identifier{
				ResourceType: meshservice_api.MeshServiceType,
				Mesh:         "default",
				SectionName:  "http",
			},
			expectErr: true,
		}),
		Entry("error empty sectionName", kriTestCase{
			id: kri.Identifier{
				ResourceType: meshservice_api.MeshServiceType,
				Mesh:         "default",
				Name:         "backend",
			},
			expectErr: true,
		}),
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
		Entry("error mesh contains dot", kriTestCase{
			id: kri.Identifier{
				ResourceType: meshservice_api.MeshServiceType,
				Mesh:         "de.fault",
				Name:         "backend",
				SectionName:  "http",
			},
			expectErr: true,
		}),
		Entry("error name contains dot", kriTestCase{
			id: kri.Identifier{
				ResourceType: meshservice_api.MeshServiceType,
				Mesh:         "default",
				Name:         "back.end",
				SectionName:  "http",
			},
			expectErr: true,
		}),
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
		Entry("error sectionName contains dot", kriTestCase{
			id: kri.Identifier{
				ResourceType: meshservice_api.MeshServiceType,
				Mesh:         "default",
				Name:         "backend",
				SectionName:  "http.port",
			},
			expectErr: true,
		}),
		Entry("label exactly 63 chars", kriTestCase{
			id: kri.Identifier{
				ResourceType: meshservice_api.MeshServiceType,
				Mesh:         "default",
				Name:         strings.Repeat("a", 63),
				SectionName:  "http",
			},
			expected: "sni.msvc.default." + strings.Repeat("a", 63) + ".http",
		}),
		Entry("error label exceeds 63 chars", kriTestCase{
			id: kri.Identifier{
				ResourceType: meshservice_api.MeshServiceType,
				Mesh:         "default",
				Name:         strings.Repeat("a", 64),
				SectionName:  "http",
			},
			expectErr: true,
		}),
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
		Entry("error namespace contains dot", kriTestCase{
			id: kri.Identifier{
				ResourceType: meshservice_api.MeshServiceType,
				Mesh:         "default",
				Zone:         "east",
				Namespace:    "app.ns",
				Name:         "backend",
				SectionName:  "http",
			},
			expectErr: true,
		}),
		Entry("error namespace exceeds 63 chars", kriTestCase{
			id: kri.Identifier{
				ResourceType: meshservice_api.MeshServiceType,
				Mesh:         "default",
				Zone:         "east",
				Namespace:    strings.Repeat("a", 64),
				Name:         "backend",
				SectionName:  "http",
			},
			expectErr: true,
		}),
		Entry("error zone exceeds 63 chars", kriTestCase{
			id: kri.Identifier{
				ResourceType: meshservice_api.MeshServiceType,
				Mesh:         "default",
				Zone:         strings.Repeat("z", 64),
				Name:         "backend",
				SectionName:  "http",
			},
			expectErr: true,
		}),
		Entry("error mesh exceeds 63 chars", kriTestCase{
			id: kri.Identifier{
				ResourceType: meshservice_api.MeshServiceType,
				Mesh:         strings.Repeat("m", 64),
				Name:         "backend",
				SectionName:  "http",
			},
			expectErr: true,
		}),
		Entry("zone exactly 63 chars", kriTestCase{
			id: kri.Identifier{
				ResourceType: meshservice_api.MeshServiceType,
				Mesh:         "default",
				Zone:         strings.Repeat("z", 63),
				Name:         "backend",
				SectionName:  "http",
			},
			expected: "sni.msvc.default." + strings.Repeat("z", 63) + ".backend.http",
		}),
		Entry("namespace exactly 63 chars", kriTestCase{
			id: kri.Identifier{
				ResourceType: meshservice_api.MeshServiceType,
				Mesh:         "default",
				Zone:         "east",
				Namespace:    strings.Repeat("n", 63),
				Name:         "backend",
				SectionName:  "http",
			},
			expected: "sni.msvc.default.east." + strings.Repeat("n", 63) + ".backend.http",
		}),
		Entry("mesh exactly 63 chars", kriTestCase{
			id: kri.Identifier{
				ResourceType: meshservice_api.MeshServiceType,
				Mesh:         strings.Repeat("m", 63),
				Name:         "backend",
				SectionName:  "http",
			},
			expected: "sni.msvc." + strings.Repeat("m", 63) + ".backend.http",
		}),
		Entry("sectionName exactly 63 chars", kriTestCase{
			id: kri.Identifier{
				ResourceType: meshservice_api.MeshServiceType,
				Mesh:         "default",
				Name:         "backend",
				SectionName:  strings.Repeat("p", 63),
			},
			expected: "sni.msvc.default.backend." + strings.Repeat("p", 63),
		}),
		Entry("error sectionName exceeds 63 chars", kriTestCase{
			id: kri.Identifier{
				ResourceType: meshservice_api.MeshServiceType,
				Mesh:         "default",
				Name:         "backend",
				SectionName:  strings.Repeat("p", 64),
			},
			expectErr: true,
		}),
		Entry("total length just under 253 chars passes", kriTestCase{
			// 4 + 1 + 4 + 1 + 60 + 1 + 60 + 1 + 60 + 1 + 60 + 1 + 4 = 258 — still > 253
			// trim to fit: 4 + 1 + 4 + 1 + 58 + 1 + 58 + 1 + 58 + 1 + 58 + 1 + 4 = 250
			id: kri.Identifier{
				ResourceType: meshservice_api.MeshServiceType,
				Mesh:         strings.Repeat("a", 58),
				Zone:         strings.Repeat("b", 58),
				Namespace:    strings.Repeat("c", 58),
				Name:         strings.Repeat("d", 58),
				SectionName:  "http",
			},
			expected: "sni.msvc." + strings.Repeat("a", 58) + "." + strings.Repeat("b", 58) + "." + strings.Repeat("c", 58) + "." + strings.Repeat("d", 58) + ".http",
		}),
		Entry("MeshExternalService zone with namespace", kriTestCase{
			id: kri.Identifier{
				ResourceType: meshexternalservice_api.MeshExternalServiceType,
				Mesh:         "prod",
				Zone:         "west",
				Namespace:    "external-ns",
				Name:         "ext-backend",
				SectionName:  "9000",
			},
			expected: "sni.extsvc.prod.west.external-ns.ext-backend.9000",
		}),
		Entry("MeshMultiZoneService zone with namespace", kriTestCase{
			id: kri.Identifier{
				ResourceType: meshmzservice_api.MeshMultiZoneServiceType,
				Mesh:         "default",
				Zone:         "east",
				Namespace:    "global-ns",
				Name:         "global-svc",
				SectionName:  "http",
			},
			expected: "sni.mzsvc.default.east.global-ns.global-svc.http",
		}),
		Entry("name with hyphens is valid", kriTestCase{
			id: kri.Identifier{
				ResourceType: meshservice_api.MeshServiceType,
				Mesh:         "default",
				Name:         "back-end-service",
				SectionName:  "http",
			},
			expected: "sni.msvc.default.back-end-service.http",
		}),
		Entry("numeric-only sectionName is valid", kriTestCase{
			id: kri.Identifier{
				ResourceType: meshservice_api.MeshServiceType,
				Mesh:         "default",
				Name:         "backend",
				SectionName:  "65535",
			},
			expected: "sni.msvc.default.backend.65535",
		}),
		Entry("error mesh contains uppercase", kriTestCase{
			id: kri.Identifier{
				ResourceType: meshservice_api.MeshServiceType,
				Mesh:         "Default",
				Name:         "backend",
				SectionName:  "http",
			},
			expectErr: true,
		}),
		Entry("error name contains underscore", kriTestCase{
			id: kri.Identifier{
				ResourceType: meshservice_api.MeshServiceType,
				Mesh:         "default",
				Name:         "back_end",
				SectionName:  "http",
			},
			expectErr: true,
		}),
		Entry("error name starts with digit", kriTestCase{
			id: kri.Identifier{
				ResourceType: meshservice_api.MeshServiceType,
				Mesh:         "default",
				Name:         "1backend",
				SectionName:  "http",
			},
			expectErr: true,
		}),
		Entry("error name starts with dash", kriTestCase{
			id: kri.Identifier{
				ResourceType: meshservice_api.MeshServiceType,
				Mesh:         "default",
				Name:         "-backend",
				SectionName:  "http",
			},
			expectErr: true,
		}),
		Entry("error name ends with dash", kriTestCase{
			id: kri.Identifier{
				ResourceType: meshservice_api.MeshServiceType,
				Mesh:         "default",
				Name:         "backend-",
				SectionName:  "http",
			},
			expectErr: true,
		}),
		Entry("error zone contains uppercase", kriTestCase{
			id: kri.Identifier{
				ResourceType: meshservice_api.MeshServiceType,
				Mesh:         "default",
				Zone:         "East",
				Name:         "backend",
				SectionName:  "http",
			},
			expectErr: true,
		}),
		Entry("error namespace contains underscore", kriTestCase{
			id: kri.Identifier{
				ResourceType: meshservice_api.MeshServiceType,
				Mesh:         "default",
				Zone:         "east",
				Namespace:    "app_ns",
				Name:         "backend",
				SectionName:  "http",
			},
			expectErr: true,
		}),
		Entry("error sectionName contains uppercase", kriTestCase{
			id: kri.Identifier{
				ResourceType: meshservice_api.MeshServiceType,
				Mesh:         "default",
				Name:         "backend",
				SectionName:  "HTTP",
			},
			expectErr: true,
		}),
		Entry("error sectionName mixes digits with letters starting with digit", kriTestCase{
			// "8080x" — not all-digits (so the numeric carve-out doesn't apply)
			// and starts with a digit (so it fails RFC 1035).
			id: kri.Identifier{
				ResourceType: meshservice_api.MeshServiceType,
				Mesh:         "default",
				Name:         "backend",
				SectionName:  "8080x",
			},
			expectErr: true,
		}),
		Entry("error sectionName numeric but too long", kriTestCase{
			// 6 digits exceeds the 5-digit (max port) numeric carve-out.
			id: kri.Identifier{
				ResourceType: meshservice_api.MeshServiceType,
				Mesh:         "default",
				Name:         "backend",
				SectionName:  "123456",
			},
			expectErr: true,
		}),
		Entry("sectionName with letters then digits is valid", kriTestCase{
			id: kri.Identifier{
				ResourceType: meshservice_api.MeshServiceType,
				Mesh:         "default",
				Name:         "backend",
				SectionName:  "http2",
			},
			expected: "sni.msvc.default.backend.http2",
		}),
	)

	It("reports an RFC 1035 violation for an over-long label", func() {
		errs := sni.ValidateKRI(kri.Identifier{
			ResourceType: meshservice_api.MeshServiceType,
			Mesh:         "default",
			Name:         strings.Repeat("a", 64),
			SectionName:  "http",
		})
		Expect(errs).ToNot(BeEmpty())
		Expect(errs[0].Error()).To(ContainSubstring("RFC 1035"))
	})

	It("reports a DNS hostname limit violation", func() {
		errs := sni.ValidateKRI(kri.Identifier{
			ResourceType: meshservice_api.MeshServiceType,
			Mesh:         strings.Repeat("a", 63),
			Zone:         strings.Repeat("b", 63),
			Namespace:    strings.Repeat("c", 63),
			Name:         strings.Repeat("d", 63),
			SectionName:  "http",
		})
		Expect(errs).ToNot(BeEmpty())
		var joined strings.Builder
		for _, e := range errs {
			joined.WriteString(e.Error() + "\n")
		}
		Expect(joined.String()).To(ContainSubstring("DNS hostname limit"))
	})

	It("reports multiple independent violations at once", func() {
		errs := sni.ValidateKRI(kri.Identifier{
			ResourceType: meshservice_api.MeshServiceType,
			Mesh:         "default",
			Name:         "foo.bar",               // dot in name
			SectionName:  strings.Repeat("a", 64), // port > 63 chars
		})
		Expect(len(errs)).To(BeNumerically(">=", 2))
	})

	It("reports separate errors for each non-conforming label", func() {
		errs := sni.ValidateKRI(kri.Identifier{
			ResourceType: meshservice_api.MeshServiceType,
			Mesh:         strings.Repeat("a", 64),
			Zone:         strings.Repeat("b", 64),
			Name:         strings.Repeat("c", 64),
			SectionName:  "http",
		})
		labelErrs := 0
		for _, e := range errs {
			if strings.Contains(e.Error(), "RFC 1035") {
				labelErrs++
			}
		}
		Expect(labelErrs).To(BeNumerically(">=", 3))
	})

	It("returns nil for an unknown resource type", func() {
		Expect(sni.ValidateKRI(kri.Identifier{
			ResourceType: model.ResourceType("DoesNotExist"),
			Mesh:         "default",
			Name:         "backend",
			SectionName:  "http",
		})).To(BeNil())
	})

	It("returns nil for a registered but non-SNI resource type", func() {
		// core_mesh.MeshResource is registered with short name "m", not in the SNI-capable set.
		Expect(sni.ValidateKRI(kri.Identifier{
			ResourceType: core_mesh.MeshType,
			Mesh:         "default",
			Name:         "backend",
			SectionName:  "http",
		})).To(BeNil())
	})

	It("emits an error message that names the offending field", func() {
		errs := sni.ValidateKRI(kri.Identifier{
			ResourceType: meshservice_api.MeshServiceType,
			Mesh:         "de.fault",
			Name:         "backend",
			SectionName:  "http",
		})
		Expect(errs).To(HaveLen(1))
		Expect(errs[0].Error()).To(ContainSubstring("mesh"))
		Expect(errs[0].Error()).To(ContainSubstring("de.fault"))
		Expect(errs[0].Error()).To(ContainSubstring("RFC 1035"))
	})

	It("flags namespace set without zone independently of length", func() {
		errs := sni.ValidateKRI(kri.Identifier{
			ResourceType: meshservice_api.MeshServiceType,
			Mesh:         "default",
			Namespace:    "ns",
			Name:         "backend",
			SectionName:  "http",
		})
		Expect(errs).To(HaveLen(1))
		Expect(errs[0].Error()).To(ContainSubstring("namespace"))
		Expect(errs[0].Error()).To(ContainSubstring("zone"))
	})
})
