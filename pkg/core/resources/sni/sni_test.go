package sni_test

import (
	"strings"

	"github.com/asaskevich/govalidator"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/v3/pkg/core/kri"
	core_mesh "github.com/kumahq/kuma/v3/pkg/core/resources/apis/mesh"
	meshexternalservice_api "github.com/kumahq/kuma/v3/pkg/core/resources/apis/meshexternalservice/api/v1alpha1"
	meshmzservice_api "github.com/kumahq/kuma/v3/pkg/core/resources/apis/meshmultizoneservice/api/v1alpha1"
	meshservice_api "github.com/kumahq/kuma/v3/pkg/core/resources/apis/meshservice/api/v1alpha1"
	"github.com/kumahq/kuma/v3/pkg/core/resources/model"
	"github.com/kumahq/kuma/v3/pkg/core/resources/sni"
)

const (
	regex1123 = "a lowercase RFC 1123 label must consist of lower case alphanumeric characters or '-', and must start and end with an alphanumeric character (e.g. 'my-name',  or '123-abc', regex used for validation is '[a-z0-9]([-a-z0-9]*[a-z0-9])?')"
	maxLen63  = "must be no more than 63 characters"
	noDots    = "must not contain dots"
	emptyMsg  = "mesh, name and sectionName must be non-empty"
)

var _ = Describe("FromKRI / ValidateKRI", func() {
	type kriTestCase struct {
		id           kri.Identifier
		expected     string
		expectedErrs []string // exact error texts; empty = expect no errors
	}
	DescribeTable("",
		func(tc kriTestCase) {
			errs := sni.ValidateKRI(tc.id)
			if len(tc.expectedErrs) > 0 {
				actual := make([]string, 0, len(errs))
				for _, e := range errs {
					actual = append(actual, e.Error())
				}
				Expect(actual).To(ConsistOf(tc.expectedErrs))
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
			expectedErrs: []string{emptyMsg},
		}),
		Entry("error empty name", kriTestCase{
			id: kri.Identifier{
				ResourceType: meshservice_api.MeshServiceType,
				Mesh:         "default",
				SectionName:  "http",
			},
			expectedErrs: []string{emptyMsg},
		}),
		Entry("error empty sectionName", kriTestCase{
			id: kri.Identifier{
				ResourceType: meshservice_api.MeshServiceType,
				Mesh:         "default",
				Name:         "backend",
			},
			expectedErrs: []string{emptyMsg},
		}),
		Entry("namespace ignored when zone is empty (global k8s)", kriTestCase{
			// Global-originated resource carries the kube namespace label but no zone,
			// so the SNI must collapse to the 5-segment global form.
			id: kri.Identifier{
				ResourceType: meshservice_api.MeshServiceType,
				Mesh:         "default",
				Namespace:    "kuma-system",
				Name:         "backend",
				SectionName:  "http",
			},
			expected: "sni.msvc.default.backend.http",
		}),
		Entry("invalid namespace ignored when zone is empty", kriTestCase{
			// A non-RFC-1035 namespace value must not produce a warning when the
			// resource is global-originated, since the segment never reaches the SNI.
			id: kri.Identifier{
				ResourceType: meshservice_api.MeshServiceType,
				Mesh:         "default",
				Namespace:    "Bad_NS",
				Name:         "backend",
				SectionName:  "http",
			},
			expected: "sni.msvc.default.backend.http",
		}),
		Entry("error mesh contains dot", kriTestCase{
			id: kri.Identifier{
				ResourceType: meshservice_api.MeshServiceType,
				Mesh:         "de.fault",
				Name:         "backend",
				SectionName:  "http",
			},
			expectedErrs: []string{`mesh "de.fault" does not conform to RFC 1123: ` + noDots},
		}),
		Entry("error name contains dot", kriTestCase{
			id: kri.Identifier{
				ResourceType: meshservice_api.MeshServiceType,
				Mesh:         "default",
				Name:         "back.end",
				SectionName:  "http",
			},
			expectedErrs: []string{`name "back.end" does not conform to RFC 1123: ` + noDots},
		}),
		Entry("error zone contains dot", kriTestCase{
			id: kri.Identifier{
				ResourceType: meshservice_api.MeshServiceType,
				Mesh:         "default",
				Zone:         "east.zone",
				Name:         "backend",
				SectionName:  "http",
			},
			expectedErrs: []string{`zone "east.zone" does not conform to RFC 1123: ` + noDots},
		}),
		Entry("error sectionName contains dot", kriTestCase{
			id: kri.Identifier{
				ResourceType: meshservice_api.MeshServiceType,
				Mesh:         "default",
				Name:         "backend",
				SectionName:  "http.port",
			},
			expectedErrs: []string{noDots},
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
			expectedErrs: []string{`name "` + strings.Repeat("a", 64) + `" does not conform to RFC 1123: ` + maxLen63},
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
			expectedErrs: []string{"computed SNI is 269 characters which exceeds the DNS hostname limit (253)"},
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
			expectedErrs: []string{`namespace "app.ns" does not conform to RFC 1123: ` + noDots},
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
			expectedErrs: []string{`namespace "` + strings.Repeat("a", 64) + `" does not conform to RFC 1123: ` + maxLen63},
		}),
		Entry("error zone exceeds 63 chars", kriTestCase{
			id: kri.Identifier{
				ResourceType: meshservice_api.MeshServiceType,
				Mesh:         "default",
				Zone:         strings.Repeat("z", 64),
				Name:         "backend",
				SectionName:  "http",
			},
			expectedErrs: []string{`zone "` + strings.Repeat("z", 64) + `" does not conform to RFC 1123: ` + maxLen63},
		}),
		Entry("error mesh exceeds 63 chars", kriTestCase{
			id: kri.Identifier{
				ResourceType: meshservice_api.MeshServiceType,
				Mesh:         strings.Repeat("m", 64),
				Name:         "backend",
				SectionName:  "http",
			},
			expectedErrs: []string{`mesh "` + strings.Repeat("m", 64) + `" does not conform to RFC 1123: ` + maxLen63},
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
			expectedErrs: []string{maxLen63},
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
		Entry("error mesh contains uppercase", kriTestCase{
			id: kri.Identifier{
				ResourceType: meshservice_api.MeshServiceType,
				Mesh:         "Default",
				Name:         "backend",
				SectionName:  "http",
			},
			expectedErrs: []string{`mesh "Default" does not conform to RFC 1123: ` + regex1123},
		}),
		Entry("error name contains underscore", kriTestCase{
			id: kri.Identifier{
				ResourceType: meshservice_api.MeshServiceType,
				Mesh:         "default",
				Name:         "back_end",
				SectionName:  "http",
			},
			expectedErrs: []string{`name "back_end" does not conform to RFC 1123: ` + regex1123},
		}),
		Entry("error name starts with dash", kriTestCase{
			id: kri.Identifier{
				ResourceType: meshservice_api.MeshServiceType,
				Mesh:         "default",
				Name:         "-backend",
				SectionName:  "http",
			},
			expectedErrs: []string{`name "-backend" does not conform to RFC 1123: ` + regex1123},
		}),
		Entry("error name ends with dash", kriTestCase{
			id: kri.Identifier{
				ResourceType: meshservice_api.MeshServiceType,
				Mesh:         "default",
				Name:         "backend-",
				SectionName:  "http",
			},
			expectedErrs: []string{`name "backend-" does not conform to RFC 1123: ` + regex1123},
		}),
		Entry("error zone contains uppercase", kriTestCase{
			id: kri.Identifier{
				ResourceType: meshservice_api.MeshServiceType,
				Mesh:         "default",
				Zone:         "East",
				Name:         "backend",
				SectionName:  "http",
			},
			expectedErrs: []string{`zone "East" does not conform to RFC 1123: ` + regex1123},
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
			expectedErrs: []string{`namespace "app_ns" does not conform to RFC 1123: ` + regex1123},
		}),
		Entry("error sectionName contains uppercase", kriTestCase{
			id: kri.Identifier{
				ResourceType: meshservice_api.MeshServiceType,
				Mesh:         "default",
				Name:         "backend",
				SectionName:  "HTTP",
			},
			expectedErrs: []string{regex1123},
		}),
	)

	DescribeTable("valid sectionName forms",
		func(sectionName string) {
			id := kri.Identifier{
				ResourceType: meshservice_api.MeshServiceType,
				Mesh:         "default",
				Name:         "backend",
				SectionName:  sectionName,
			}
			Expect(sni.ValidateKRI(id)).To(BeEmpty())
			Expect(sni.FromKRI(id)).To(Equal("sni.msvc.default.backend." + sectionName))
		},
		Entry("numeric only", "65535"),
		Entry("six digits", "123456"),
		Entry("digits then letter", "8080x"),
		Entry("letters then digits", "http2"),
	)

	It("reports an RFC 1123 violation for an over-long label", func() {
		longName := strings.Repeat("a", 64)
		errs := sni.ValidateKRI(kri.Identifier{
			ResourceType: meshservice_api.MeshServiceType,
			Mesh:         "default",
			Name:         longName,
			SectionName:  "http",
		})
		Expect(errs).To(HaveLen(1))
		Expect(errs[0]).To(MatchError(`name "` + longName + `" does not conform to RFC 1123: ` + maxLen63))
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
		Expect(errs).To(HaveLen(1))
		Expect(errs[0]).To(MatchError(`computed SNI is 269 characters which exceeds the DNS hostname limit (253)`))
	})

	It("reports multiple independent violations at once", func() {
		errs := sni.ValidateKRI(kri.Identifier{
			ResourceType: meshservice_api.MeshServiceType,
			Mesh:         "default",
			Name:         "foo.bar",               // dot in name
			SectionName:  strings.Repeat("a", 64), // port > 63 chars
		})
		Expect(errs).To(HaveLen(2))
		Expect(errs[0]).To(MatchError(`name "foo.bar" does not conform to RFC 1123: ` + noDots))
		Expect(errs[1]).To(MatchError(maxLen63))
	})

	It("reports separate errors for each non-conforming label", func() {
		errs := sni.ValidateKRI(kri.Identifier{
			ResourceType: meshservice_api.MeshServiceType,
			Mesh:         strings.Repeat("a", 64),
			Zone:         strings.Repeat("b", 64),
			Name:         strings.Repeat("c", 64),
			SectionName:  "http",
		})
		actual := make([]string, 0, len(errs))
		for _, e := range errs {
			actual = append(actual, e.Error())
		}
		Expect(actual).To(ContainElements(
			`mesh "`+strings.Repeat("a", 64)+`" does not conform to RFC 1123: `+maxLen63,
			`zone "`+strings.Repeat("b", 64)+`" does not conform to RFC 1123: `+maxLen63,
			`name "`+strings.Repeat("c", 64)+`" does not conform to RFC 1123: `+maxLen63,
		))
	})

	It("surfaces an error for an unknown resource type", func() {
		errs := sni.ValidateKRI(kri.Identifier{
			ResourceType: model.ResourceType("DoesNotExist"),
			Mesh:         "default",
			Name:         "backend",
			SectionName:  "http",
		})
		Expect(errs).To(HaveLen(1))
		Expect(errs[0].Error()).To(HavePrefix(`unknown resource type "DoesNotExist"`))
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
		Expect(errs[0]).To(MatchError(`mesh "de.fault" does not conform to RFC 1123: ` + noDots))
	})
})
